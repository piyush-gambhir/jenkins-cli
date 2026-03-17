package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newViewDeleteCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <view-name>",
		Short: "Delete a view",
		Long: `Delete a Jenkins view.

Removes the view from the Jenkins dashboard. This does not delete the
jobs contained in the view. Requires --confirm.

Examples:
  # Delete a view
  jenkins view delete "My View" --confirm`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !confirm {
				return fmt.Errorf("use --confirm to confirm deletion of view %q", name)
			}

			if err := jenkinsClient.DeleteView(name); err != nil {
				return fmt.Errorf("deleting view: %w", err)
			}

			fmt.Fprintf(os.Stdout, "View %q deleted.\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm deletion")

	return cmd
}
