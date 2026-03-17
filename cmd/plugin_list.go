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

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		Long:  "List all installed Jenkins plugins.",
		Args:  cobra.NoArgs,
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

	cmd.Flags().BoolVar(&activeOnly, "active", false, "Show only active/enabled plugins")

	return cmd
}
