package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/config"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
	"github.com/piyush-gambhir/jenkins-cli/internal/update"
	"github.com/piyush-gambhir/jenkins-cli/internal/version"
)

var (
	// Global flags
	outputFormat string
	profileFlag  string
	serverFlag   string
	userFlag     string
	tokenFlag    string
	insecureFlag bool
	noColorFlag  bool
	verboseFlag  bool
	readOnlyFlag bool
	noInputFlag  bool
	quietFlag    bool

	// Shared state set during PersistentPreRunE
	cfg           *config.Config
	jenkinsClient *client.Client
	outFormat     output.Format

	// OutputFormat is the exported package-level output format string, set
	// during PersistentPreRunE so that main.go error handling can use it.
	OutputFormat string

	// Update check channel (replaces sync.Mutex pattern)
	updateResult chan *update.UpdateInfo
)

var rootCmd = &cobra.Command{
	Use:   "jenkins",
	Short: "Jenkins CLI — manage Jenkins from the command line",
	Long: `A comprehensive command-line interface for interacting with Jenkins CI/CD servers.

Manage jobs, builds, nodes, plugins, credentials, pipelines, views, and
system administration from the terminal. Designed for both human operators
and coding agents (LLMs).

Quick start:
  jenkins login                           # authenticate with a Jenkins server
  jenkins status                          # check server connectivity
  jenkins job list                        # list all jobs
  jenkins job build my-pipeline --follow  # trigger a build and stream logs

All list/get commands support -o json and -o yaml for machine-readable output.

Use "jenkins <command> --help" for detailed information about any command.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Check env vars for --no-input and --quiet
		if !noInputFlag {
			if v, ok := os.LookupEnv("JENKINS_NO_INPUT"); ok && v != "" && v != "0" && v != "false" {
				noInputFlag = true
			}
		}
		if !quietFlag {
			if v, ok := os.LookupEnv("JENKINS_QUIET"); ok && v != "" && v != "0" && v != "false" {
				quietFlag = true
			}
		}

		// Start background update check for commands that should show it
		cmdName := cmd.Name()
		if cmdName != "update" && cmdName != "version" {
			startBackgroundUpdateCheck()
		}

		// Skip auth for commands that don't need it
		if cmdName == "version" || cmdName == "help" || cmdName == "login" || cmdName == "update" {
			return nil
		}
		// Also skip for parent commands (they have subcommands)
		if cmd.HasSubCommands() && cmd.Args == nil {
			return nil
		}

		if err := loadConfig(cmd); err != nil {
			return err
		}

		profile, err := resolveProfile(cmd)
		if err != nil {
			return err
		}

		if err := setupJenkinsClient(&profile); err != nil {
			return err
		}

		if err := checkPermissions(cmd, &profile); err != nil {
			return err
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// Wait for background update check and print notice if available.
		// Skip for update and version commands.
		cmdName := cmd.Name()
		if cmdName == "update" || cmdName == "version" {
			return nil
		}

		if updateResult == nil {
			return nil
		}

		select {
		case info := <-updateResult:
			if info != nil && info.Available {
				update.PrintUpdateNotice(os.Stderr, info)
			}
		case <-time.After(2 * time.Second):
			// Don't block the user if the update check is slow
		}

		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// startBackgroundUpdateCheck launches a goroutine to check for updates using
// the 24h cache so it doesn't slow down normal command execution.
func startBackgroundUpdateCheck() {
	updateResult = make(chan *update.UpdateInfo, 1)
	go func() {
		info, _ := update.CheckForUpdate(
			version.Version,
			"piyush-gambhir/jenkins-cli",
			config.ConfigDir(),
			false,
		)
		updateResult <- info
	}()
}

// loadConfig loads the configuration and parses the output format.
func loadConfig(cmd *cobra.Command) error {
	var err error

	// Parse output format
	outFormat, err = output.ParseFormat(outputFormat)
	if err != nil {
		return err
	}

	// Store in exported package-level var for main.go error handling
	OutputFormat = outputFormat

	// Load config
	cfg, err = config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// If output format not set via flag, try config default
	if outputFormat == "" && cfg.Defaults.Output != "" {
		outFormat, err = output.ParseFormat(cfg.Defaults.Output)
		if err != nil {
			return err
		}
		OutputFormat = cfg.Defaults.Output
	}

	return nil
}

// resolveProfile resolves auth credentials from flags, env, and config.
func resolveProfile(cmd *cobra.Command) (config.Profile, error) {
	flags := config.FlagValues{
		Server:      serverFlag,
		User:        userFlag,
		Token:       tokenFlag,
		Insecure:    insecureFlag,
		ServerSet:   cmd.Flags().Changed("server"),
		UserSet:     cmd.Flags().Changed("user"),
		TokenSet:    cmd.Flags().Changed("token"),
		InsecureSet: cmd.Flags().Changed("insecure"),
	}

	profile, err := config.ResolveAuth(flags, os.LookupEnv, cfg, profileFlag)
	if err != nil {
		return config.Profile{}, fmt.Errorf("resolving auth: %w", err)
	}

	if profile.URL == "" {
		return config.Profile{}, fmt.Errorf("Jenkins URL not configured. Run 'jenkins login' or set JENKINS_URL")
	}

	return profile, nil
}

// setupJenkinsClient creates the Jenkins API client from the resolved profile.
func setupJenkinsClient(profile *config.Profile) error {
	jenkinsClient = client.NewClient(*profile, verboseFlag)
	return nil
}

// checkPermissions enforces read-only mode and no-input restrictions.
func checkPermissions(cmd *cobra.Command, profile *config.Profile) error {
	effectiveReadOnly := profile.ReadOnly
	if cmd.Flags().Changed("read-only") {
		effectiveReadOnly = readOnlyFlag
	}
	if effectiveReadOnly && cmd.Annotations != nil && cmd.Annotations["mutates"] == "true" {
		return fmt.Errorf("command '%s' is blocked in read-only mode.\nTo disable, use --read-only=false or remove read_only from your config profile.", cmd.CommandPath())
	}

	return nil
}

// RootCmd returns the root cobra.Command for use in main.go.
func RootCmd() *cobra.Command {
	return rootCmd
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		statusCode := 0
		var apiErr *client.APIError
		if errors.As(err, &apiErr) {
			statusCode = apiErr.StatusCode
		}
		output.WriteError(os.Stderr, outFormat, err, statusCode)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().StringVar(&profileFlag, "profile", "", "Configuration profile to use")
	rootCmd.PersistentFlags().StringVarP(&serverFlag, "server", "s", "", "Jenkins server URL")
	rootCmd.PersistentFlags().StringVarP(&userFlag, "user", "u", "", "Jenkins username")
	rootCmd.PersistentFlags().StringVarP(&tokenFlag, "token", "t", "", "Jenkins API token")
	rootCmd.PersistentFlags().BoolVarP(&insecureFlag, "insecure", "k", false, "Skip TLS verification")
	rootCmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&readOnlyFlag, "read-only", false, "Block write operations (safety mode for agents)")
	rootCmd.PersistentFlags().BoolVar(&noInputFlag, "no-input", false, "Disable all interactive prompts (for CI/agent use)")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Suppress informational output")

	// Register all subcommands
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newLoginCmd())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newWhoAmICmd())
	rootCmd.AddCommand(newJobCmd())
	rootCmd.AddCommand(newBuildCmd())
	rootCmd.AddCommand(newQueueCmd())
	rootCmd.AddCommand(newNodeCmd())
	rootCmd.AddCommand(newViewCmd())
	rootCmd.AddCommand(newPluginCmd())
	rootCmd.AddCommand(newCredentialCmd())
	rootCmd.AddCommand(newUserCmd())
	rootCmd.AddCommand(newPipelineCmd())
	rootCmd.AddCommand(newSystemCmd())
	rootCmd.AddCommand(newUpdateCmd())
}
