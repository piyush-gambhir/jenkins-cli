package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newPluginListCmd() *cobra.Command {
	var activeOnly bool
	var enabledOnly bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		Long: `List all installed Jenkins plugins.

Displays each plugin's short name, version, enabled status, and whether
an update is available. Use --active to show only plugins that are both
active and enabled. Use --enabled to show only enabled plugins (which
may include inactive ones).

Examples:
  # List all installed plugins
  jenkins plugin list

  # List only active and enabled plugins
  jenkins plugin list --active

  # List only enabled plugins
  jenkins plugin list --enabled

  # Output as JSON for scripting
  jenkins plugin list -o json

  # Output as YAML
  jenkins plugin list -o yaml`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			plugins, err := jenkinsClient.ListPlugins()
			if err != nil {
				return fmt.Errorf("listing plugins: %w", err)
			}

			if activeOnly {
				var filtered []client.Plugin
				for _, p := range plugins {
					if p.Active && p.Enabled {
						filtered = append(filtered, p)
					}
				}
				plugins = filtered
			} else if enabledOnly {
				var filtered []client.Plugin
				for _, p := range plugins {
					if p.Enabled {
						filtered = append(filtered, p)
					}
				}
				plugins = filtered
			}

			if len(plugins) == 0 {
				fmt.Fprintln(os.Stdout, "No plugins found.")
				return nil
			}

			tableDef := &output.TableDef{
				Headers: []string{"SHORT NAME", "VERSION", "ENABLED", "HAS UPDATE"},
				RowFunc: func(item interface{}) []string {
					p := item.(client.Plugin)
					return []string{
						p.ShortName,
						p.Version,
						fmt.Sprintf("%v", p.Enabled),
						fmt.Sprintf("%v", p.HasUpdate),
					}
				},
			}

			return output.Print(os.Stdout, outFormat, plugins, tableDef)
		},
	}

	cmd.Flags().BoolVar(&activeOnly, "active", false, "Show only active and enabled plugins")
	cmd.Flags().BoolVar(&enabledOnly, "enabled", false, "Show only enabled plugins")

	return cmd
}
