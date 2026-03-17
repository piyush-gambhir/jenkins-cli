package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newViewGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <view-name>",
		Short: "Get view details",
		Long:  "Display detailed information about a view including its jobs.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			view, err := jenkinsClient.GetView(name)
			if err != nil {
				return fmt.Errorf("getting view: %w", err)
			}

			if outFormat == output.FormatTable {
				fmt.Fprintf(os.Stdout, "View: %s\n", view.Name)
				fmt.Fprintf(os.Stdout, "  URL:         %s\n", view.URL)
				if view.Description != "" {
					fmt.Fprintf(os.Stdout, "  Description: %s\n", view.Description)
				}
				if len(view.Jobs) > 0 {
					fmt.Fprintf(os.Stdout, "  Jobs (%d):\n", len(view.Jobs))
					for _, j := range view.Jobs {
						fmt.Fprintf(os.Stdout, "    - %s (%s)\n", j.Name, client.ColorToStatus(j.Color))
					}
				}
				return nil
			}

			return output.Print(os.Stdout, outFormat, view, nil)
		},
	}
}
