package cmd

import (
	"fmt"
	"os"
	"sync"

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

	// Shared state set during PersistentPreRunE
	cfg           *config.Config
	jenkinsClient *client.Client
	outFormat     output.Format

	// Update check state
	updateInfo   *update.UpdateInfo
	updateInfoMu sync.Mutex
	updateWg     sync.WaitGroup
)

var rootCmd = &cobra.Command{
	Use:   "jenkins",
	Short: "Jenkins CLI — manage Jenkins from the command line",
	Long:  "A comprehensive command-line interface for interacting with Jenkins CI/CD servers.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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

		var err error

		// Parse output format
		outFormat, err = output.ParseFormat(outputFormat)
		if err != nil {
			return err
		}

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
		}

		// Resolve auth
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
			return fmt.Errorf("resolving auth: %w", err)
		}

		if profile.URL == "" {
			return fmt.Errorf("Jenkins URL not configured. Run 'jenkins login' or set JENKINS_URL")
		}

		jenkinsClient = client.NewClient(profile)
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// Wait for background update check and print notice if available.
		// Skip for update and version commands.
		cmdName := cmd.Name()
		if cmdName == "update" || cmdName == "version" {
			return nil
		}

		updateWg.Wait()

		updateInfoMu.Lock()
		info := updateInfo
		updateInfoMu.Unlock()

		if info != nil && info.Available {
			update.PrintUpdateNotice(os.Stderr, info)
		}

		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// startBackgroundUpdateCheck launches a goroutine to check for updates using
// the 24h cache so it doesn't slow down normal command execution.
func startBackgroundUpdateCheck() {
	updateWg.Add(1)
	go func() {
		defer updateWg.Done()

		info, err := update.CheckForUpdate(
			version.Version,
			"piyush-gambhir/jenkins-cli",
			config.ConfigDir(),
			false,
		)
		if err != nil {
			// Silently ignore update check failures
			return
		}

		updateInfoMu.Lock()
		updateInfo = info
		updateInfoMu.Unlock()
	}()
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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
