package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/update"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long: `Print the version, commit, and build time.

When an earlier update check cached the latest release, also print
"latest" and "update_available". This command never contacts GitHub.`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, version.Info())
			if info := update.CachedUpdate(version.Version, config.ConfigDir()); info != nil {
				fmt.Fprintf(out, "latest: %s\nupdate_available: %t\n", info.LatestVersion, info.Available)
			}
		},
	}
}
