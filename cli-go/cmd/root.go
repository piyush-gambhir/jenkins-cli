package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/output"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/update"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/version"
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

	// updateResult receives the update check result for this run; nil when
	// the notifier is suppressed.
	updateResult chan *update.UpdateInfo
	// updateCheckStarted is true when this run asked GitHub (the cache was
	// stale), so PostRun may wait updateCheckGrace for the answer.
	updateCheckStarted bool
	// updateCheckGrace bounds that wait. The check is recorded before the
	// request, so the wait happens at most once a day.
	updateCheckGrace = time.Second
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

Use "jenkins <command> --help" for detailed information about any command.

Full command reference (for agents/LLMs): https://jenkins-cli.pages.dev/llms.txt
Claude Code skill: https://github.com/piyush-gambhir/jenkins-cli/blob/main/jenkins/SKILL.md`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Check env vars for --no-input and --quiet
		if envFlagEnabled("JENKINS_NO_INPUT") {
			noInputFlag = true
		}
		if !cmd.Flags().Changed("quiet") {
			quietFlag = envFlagEnabled("JENKINS_QUIET")
		}

		if runtime.GOOS == "windows" {
			if execPath, err := update.ExecutablePath(); err == nil {
				update.RemoveOldBinary(runtime.GOOS, execPath)
			}
		}

		// Match on the command directly under the root, so `job update` is
		// not mistaken for `update`.
		topName := topLevelName(cmd)
		startUpdateCheck(topName)

		// Skip auth for commands that don't need it
		if skipsAuth(topName) {
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

		if err := setupJenkinsClient(cmd.Context(), &profile); err != nil {
			return err
		}

		if err := checkPermissions(cmd, &profile); err != nil {
			return err
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		printUpdateNotice(cmd.ErrOrStderr())
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func envFlagEnabled(name string) bool {
	v := strings.TrimSpace(os.Getenv(name))
	return strings.EqualFold(v, "true") || v == "1"
}

// topLevelName returns the name of the command directly under the root.
func topLevelName(cmd *cobra.Command) string {
	for cmd.HasParent() && cmd.Parent().HasParent() {
		cmd = cmd.Parent()
	}
	return cmd.Name()
}

// skipsAuth reports whether a top-level command runs without a Jenkins
// server: it must not need config, credentials, or the client.
func skipsAuth(topName string) bool {
	switch topName {
	case "version", "help", "login", "update", "completion":
		return true
	}
	return strings.HasPrefix(topName, "__complete")
}

// stderrIsTerminal is a seam for tests.
var stderrIsTerminal = func() bool { return term.IsTerminal(int(os.Stderr.Fd())) }

// startUpdateCheck prepares the update notice for this run, unless the
// notifier is suppressed (then nothing touches the cache or the network). A
// fresh cached result is used directly; otherwise GitHub is queried in the
// background while the command runs.
func startUpdateCheck(topName string) {
	updateResult, updateCheckStarted = nil, false
	if !update.NotifierEnabled(os.Getenv, stderrIsTerminal(), quietFlag, version.Version, topName) {
		return
	}
	result := make(chan *update.UpdateInfo, 1)
	updateResult = result
	current, configDir := version.Version, config.ConfigDir()
	if info, ok := update.FreshCache(current, configDir); ok {
		result <- info
		return
	}
	check := checkForUpdate
	updateCheckStarted = true
	go func() {
		info, _ := check(current, configDir, false)
		result <- info
	}()
}

// printUpdateNotice prints the update notice if it is due. A result from the
// cache is used without waiting. A check this run started gets at most
// updateCheckGrace to answer (it has its own 3s timeout); without that grace a
// fast command would exit first and lose the day's check.
func printUpdateNotice(w io.Writer) {
	if updateResult == nil {
		return
	}
	var info *update.UpdateInfo
	if updateCheckStarted {
		timer := time.NewTimer(updateCheckGrace)
		defer timer.Stop()
		select {
		case info = <-updateResult:
		case <-timer.C:
			return
		}
	} else {
		select {
		case info = <-updateResult:
		default:
			return
		}
	}
	if info != nil && info.Available {
		update.MaybeNotify(w, info, config.ConfigDir(), installMethod(), time.Now())
	}
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
func setupJenkinsClient(ctx context.Context, profile *config.Profile) error {
	jenkinsClient = client.NewClient(*profile, verboseFlag)
	jenkinsClient.SetContext(ctx)
	return nil
}

// checkPermissions enforces read-only mode and no-input restrictions.
func checkPermissions(cmd *cobra.Command, profile *config.Profile) error {
	effectiveReadOnly := profile.ReadOnly
	if readOnlyFlag {
		effectiveReadOnly = true
	}
	if effectiveReadOnly && cmd.Annotations != nil && cmd.Annotations["mutates"] == "true" {
		return readOnlyError(cmd)
	}

	return nil
}

func readOnlyError(cmd *cobra.Command) error {
	return fmt.Errorf("command '%s' is blocked in read-only mode; remove read_only from the profile or disable the read-only environment setting to permit writes", cmd.CommandPath())
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
