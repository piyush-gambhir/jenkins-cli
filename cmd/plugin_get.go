package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newPluginGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <plugin-name>",
		Short: "Get plugin details",
		Long: `Display detailed information about an installed plugin.

Shows the plugin's full name, version, active/enabled status, pinned
status, URL, backup version, and dependencies.

Examples:
  # Get details about the git plugin
  jenkins plugin get git

  # Get details about the pipeline plugin
  jenkins plugin get workflow-aggregator

  # Output as JSON
  jenkins plugin get git -o json`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			plugin, err := jenkinsClient.GetPlugin(name)
			if err != nil {
				return fmt.Errorf("getting plugin: %w", err)
			}

			if outFormat == output.FormatTable {
				fmt.Fprintf(os.Stdout, "Plugin: %s\n", plugin.LongName)
				fmt.Fprintf(os.Stdout, "  Short Name:  %s\n", plugin.ShortName)
				fmt.Fprintf(os.Stdout, "  Version:     %s\n", plugin.Version)
				fmt.Fprintf(os.Stdout, "  Active:      %v\n", plugin.Active)
				fmt.Fprintf(os.Stdout, "  Enabled:     %v\n", plugin.Enabled)
				fmt.Fprintf(os.Stdout, "  Has Update:  %v\n", plugin.HasUpdate)
				fmt.Fprintf(os.Stdout, "  Pinned:      %v\n", plugin.Pinned)
				if plugin.URL != "" {
					fmt.Fprintf(os.Stdout, "  URL:         %s\n", plugin.URL)
				}
				if plugin.BackupVersion != "" {
					fmt.Fprintf(os.Stdout, "  Backup Ver:  %s\n", plugin.BackupVersion)
				}
				if len(plugin.Dependencies) > 0 {
					fmt.Fprintf(os.Stdout, "  Dependencies:\n")
					for _, d := range plugin.Dependencies {
						opt := ""
						if d.Optional {
							opt = " (optional)"
						}
						fmt.Fprintf(os.Stdout, "    - %s@%s%s\n", d.ShortName, d.Version, opt)
					}
				}
				return nil
			}

			return output.Print(os.Stdout, outFormat, plugin, nil)
		},
	}
}
