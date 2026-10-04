package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/output"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/update"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/version"
)

type releaseInstaller interface {
	Install(ctx context.Context, version string) error
}

// Seams for tests: they keep update off the GitHub API, the real executable,
// and the real terminal.
var (
	checkForUpdate  = update.CheckForUpdate
	installMethod   = update.CurrentInstallMethod
	executablePath  = update.ExecutablePath
	newInstaller    = func(execPath string) releaseInstaller { return update.NewInstaller(execPath) }
	stdinIsTerminal = func() bool { return term.IsTerminal(int(os.Stdin.Fd())) }
)

// updateCheckResult is the `update --check` output.
type updateCheckResult struct {
	CurrentVersion  string `json:"current_version" yaml:"current_version"`
	LatestVersion   string `json:"latest_version" yaml:"latest_version"`
	UpdateAvailable bool   `json:"update_available" yaml:"update_available"`
	ReleaseURL      string `json:"release_url" yaml:"release_url"`
	InstallMethod   string `json:"install_method" yaml:"install_method"`
}

func newUpdateCmd() *cobra.Command {
	var checkOnly, yes bool

	cmd := &cobra.Command{
		Use:         "update",
		Annotations: map[string]string{"mutates": "true"},
		Short:       "Update jenkins to the latest version",
		Long: `Check for and install the latest release of the Jenkins CLI from GitHub
Releases on macOS, Linux, and Windows.

The release archive is verified against the release's checksums.txt (SHA-256)
before the running binary is replaced. Any failure leaves the current binary
in place. If the binary's directory is not writable, re-run with sudo (or as
Administrator on Windows), or reinstall into a directory you can write to.
A binary in a Go bin directory ($GOBIN, $GOPATH/bin, ~/go/bin) was built from
source and is not replaced; update it with "git pull && make install" in your
jenkins-cli/cli-go checkout instead.

--check always queries GitHub and only reports; -o json prints
current_version, latest_version, update_available, release_url, and
install_method (self or go). --read-only blocks installing but allows --check.

Update notice: in an interactive terminal, other commands check GitHub at
most once a day in the background and print a short notice on stderr when a
new release exists (at most once a day per release). The check is skipped
when stderr is not a terminal, CI is set, JENKINS_NO_UPDATE_NOTIFIER or
NO_UPDATE_NOTIFIER is set, or --quiet / JENKINS_QUIET is on.

Examples:
  jenkins update               # ask, then install the latest release
  jenkins update --yes         # install without asking
  jenkins update --check       # report only
  jenkins update --check -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			format, err := output.ParseFormat(outputFormat)
			if err != nil {
				return err
			}
			// PersistentPreRunE skips loadConfig for update; main.go reads this
			// to format errors.
			OutputFormat = outputFormat

			current := update.NormalizeVersion(version.Version)
			if !update.IsReleaseVersion(current) {
				return fmt.Errorf("cannot update a dev build (version %q); install a release from https://github.com/%s/releases", version.Version, update.Repo)
			}
			if !checkOnly && readOnlyEnabled() {
				return readOnlyError(cmd)
			}

			info, err := checkForUpdate(current, config.ConfigDir(), true)
			if err != nil {
				return fmt.Errorf("checking for updates: %w", err)
			}
			method := installMethod()

			if checkOnly {
				return printUpdateCheck(out, format, info, method)
			}
			if !info.Available {
				fmt.Fprintf(out, "jenkins v%s is already the latest version.\n", current)
				return nil
			}

			fmt.Fprintf(out, "Update available: v%s -> v%s\n", current, info.LatestVersion)
			if method == update.InstallGo {
				fmt.Fprintf(out, "This jenkins was built from source into a Go bin directory, so it is not replaced in place.\nUpdate with: %s\nRelease notes: %s\n", update.SourceUpdateCommand, info.ReleaseURL)
				return nil
			}
			if !yes {
				if noInputFlag {
					return errors.New("update needs confirmation and --no-input is set: pass --yes to install")
				}
				if !stdinIsTerminal() {
					return errors.New("update needs confirmation and stdin is not a terminal: pass --yes to install")
				}
				fmt.Fprint(cmd.ErrOrStderr(), "Update now? [Y/n] ")
				answer, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				switch strings.ToLower(strings.TrimSpace(answer)) {
				case "", "y", "yes":
				default:
					fmt.Fprintln(out, "Update cancelled.")
					return nil
				}
			}

			execPath, err := executablePath()
			if err != nil {
				return err
			}
			if !quietFlag {
				fmt.Fprintf(cmd.ErrOrStderr(), "Downloading v%s...\n", info.LatestVersion)
			}
			if err := newInstaller(execPath).Install(cmd.Context(), info.LatestVersion); err != nil {
				return fmt.Errorf("installing v%s: %w", info.LatestVersion, err)
			}
			update.ClearCache(config.ConfigDir())
			fmt.Fprintf(out, "Updated jenkins v%s -> v%s\nRelease notes: %s\n", current, info.LatestVersion, info.ReleaseURL)
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check if an update is available, don't install")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Install without asking for confirmation")

	return cmd
}

func printUpdateCheck(w io.Writer, format output.Format, info *update.UpdateInfo, method string) error {
	result := updateCheckResult{
		CurrentVersion:  info.CurrentVersion,
		LatestVersion:   info.LatestVersion,
		UpdateAvailable: info.Available,
		ReleaseURL:      info.ReleaseURL,
		InstallMethod:   method,
	}
	if format != output.FormatTable {
		return output.Print(w, format, result, nil)
	}
	available := "no"
	if info.Available {
		available = "yes"
	}
	fmt.Fprintf(w, "Current version:  v%s\nLatest version:   v%s\nUpdate available: %s\nRelease notes:    %s\n",
		info.CurrentVersion, info.LatestVersion, available, info.ReleaseURL)
	if info.Available {
		updateCmd := "jenkins update"
		if method == update.InstallGo {
			updateCmd = update.SourceUpdateCommand
		}
		fmt.Fprintf(w, "Update with:      %s\n", updateCmd)
	}
	return nil
}

// readOnlyEnabled resolves read-only mode for commands that skip the normal
// auth bootstrap: the --read-only flag, JENKINS_READ_ONLY, or the profile.
func readOnlyEnabled() bool {
	if readOnlyFlag {
		return true
	}
	c, err := config.Load()
	if err != nil {
		c = &config.Config{}
	}
	profile, err := config.ResolveAuth(config.FlagValues{}, os.LookupEnv, c, profileFlag)
	if err != nil {
		profile, _ = config.ResolveAuth(config.FlagValues{}, os.LookupEnv, &config.Config{}, "")
	}
	return profile.ReadOnly
}
