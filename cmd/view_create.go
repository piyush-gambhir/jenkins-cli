package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newViewCreateCmd() *cobra.Command {
	var viewType string

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

  # Create a view with a specific type
  jenkins view create "Dashboard" --type hudson.model.ListView`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := jenkinsClient.CreateView(name, viewType); err != nil {
				return fmt.Errorf("creating view: %w", err)
			}

			fmt.Fprintf(os.Stdout, "View %q created.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&viewType, "type", "hudson.model.ListView", "View type class name")

	return cmd
}
