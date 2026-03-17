package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newPluginCheckUpdatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check-updates",
		Short: "Check for plugin updates",
		Long: `Check for available plugin updates.

Triggers an update check against the Jenkins update center and lists
all plugins that have newer versions available.

Examples:
  # Check for plugin updates
  jenkins plugin check-updates

  # Output as JSON
  jenkins plugin check-updates -o json`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			updates, err := jenkinsClient.CheckPluginUpdates()
			if err != nil {
				return fmt.Errorf("checking updates: %w", err)
			}

			if len(updates) == 0 {
				fmt.Fprintln(os.Stdout, "All plugins are up to date.")
				return nil
			}

			tableDef := &output.TableDef{
				Headers: []string{"PLUGIN", "CURRENT VERSION", "HAS UPDATE"},
				RowFunc: func(item interface{}) []string {
					p := item.(client.Plugin)
					return []string{
						p.ShortName,
						p.Version,
						fmt.Sprintf("%v", p.HasUpdate),
					}
				},
			}

			fmt.Fprintf(os.Stdout, "%d plugin(s) have updates available:\n", len(updates))
			return output.Print(os.Stdout, outFormat, updates, tableDef)
		},
	}
}
