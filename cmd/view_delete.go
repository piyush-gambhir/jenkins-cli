package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newViewDeleteCmd() *cobra.Command {
	var confirm bool
	var ifExists bool

	cmd := &cobra.Command{
		Use:         "delete <view-name>",
		Short:       "Delete a view",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Delete a Jenkins view.

Removes the view from the Jenkins dashboard. This does not delete the
jobs contained in the view. Requires --confirm.

Examples:
  # Delete a view
  jenkins view delete "My View" --confirm

  # Idempotent delete (no error if view doesn't exist)
  jenkins view delete "My View" --confirm --if-exists`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !confirm {
				if noInputFlag {
					return fmt.Errorf("interactive input required but --no-input is set. Use --confirm for destructive operations.")
				}
				return fmt.Errorf("use --confirm to confirm deletion of view %q", name)
			}

			if err := jenkinsClient.DeleteView(name); err != nil {
				var apiErr *client.APIError
				if ifExists && errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
					if !quietFlag {
						fmt.Fprintf(os.Stdout, "View %q does not exist, skipping.\n", name)
					}
					return nil
				}
				return fmt.Errorf("deleting view: %w", err)
			}

			if !quietFlag {
				fmt.Fprintf(os.Stdout, "View %q deleted.\n", name)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm deletion")
	cmd.Flags().BoolVar(&ifExists, "if-exists", false, "Don't error if the view doesn't exist")

	return cmd
}
