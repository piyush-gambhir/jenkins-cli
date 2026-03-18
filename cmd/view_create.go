package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newViewCreateCmd() *cobra.Command {
	var viewType string
	var ifNotExists bool

	cmd := &cobra.Command{
		Use:         "create <view-name>",
		Short:       "Create a new view",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Create a new Jenkins view.

Creates an empty view. Use --type to specify the view type class name.
After creation, use "jenkins view add-job" to add jobs to the view.

Examples:
  # Create a list view (default type)
  jenkins view create "My Team"

  # Idempotent create (no error if view already exists)
  jenkins view create "My Team" --if-not-exists`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := jenkinsClient.CreateView(name, viewType); err != nil {
				var apiErr *client.APIError
				if ifNotExists && errors.As(err, &apiErr) && (apiErr.StatusCode == 400 || apiErr.StatusCode == 409) {
					if !quietFlag {
						fmt.Fprintf(os.Stdout, "View %q already exists, skipping.\n", name)
					}
					return nil
				}
				return fmt.Errorf("creating view: %w", err)
			}

			if !quietFlag {
				fmt.Fprintf(os.Stdout, "View %q created.\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&viewType, "type", "hudson.model.ListView", "View type class name")
	cmd.Flags().BoolVar(&ifNotExists, "if-not-exists", false, "Don't error if the view already exists")

	return cmd
}
