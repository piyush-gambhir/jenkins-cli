package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newViewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List views",
		Long: `List all Jenkins views.

Displays each view's name, URL, and description.

Examples:
  # List all views
  jenkins view list

  # Output as JSON
  jenkins view list -o json`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			views, err := jenkinsClient.ListViews()
			if err != nil {
				return fmt.Errorf("listing views: %w", err)
			}

			if len(views) == 0 {
				fmt.Fprintln(os.Stdout, "No views found.")
				return nil
			}

			tableDef := &output.TableDef{
				Headers: []string{"NAME", "URL", "DESCRIPTION"},
				RowFunc: func(item interface{}) []string {
					v := item.(client.View)
					desc := v.Description
					if len(desc) > 50 {
						desc = desc[:50] + "..."
					}
					return []string{v.Name, v.URL, desc}
				},
			}

			return output.Print(os.Stdout, outFormat, views, tableDef)
		},
	}
}
