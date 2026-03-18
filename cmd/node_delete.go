package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newNodeDeleteCmd() *cobra.Command {
	var confirm bool
	var ifExists bool

	cmd := &cobra.Command{
		Use:         "delete <node-name>",
		Short:       "Delete a node",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Permanently delete a Jenkins node/agent.

WARNING: This operation is irreversible. Requires --confirm.

Examples:
  # Delete a node
  jenkins node delete my-agent --confirm

  # Idempotent delete (no error if node doesn't exist)
  jenkins node delete my-agent --confirm --if-exists`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !confirm {
				if noInputFlag {
					return fmt.Errorf("interactive input required but --no-input is set. Use --confirm for destructive operations.")
				}
				return fmt.Errorf("use --confirm to confirm deletion of node %q", name)
			}

			if err := jenkinsClient.DeleteNode(name); err != nil {
				var apiErr *client.APIError
				if ifExists && errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
					if !quietFlag {
						fmt.Fprintf(os.Stdout, "Node %q does not exist, skipping.\n", name)
					}
					return nil
				}
				return fmt.Errorf("deleting node: %w", err)
			}

			if !quietFlag {
				fmt.Fprintf(os.Stdout, "Node %q deleted.\n", name)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm deletion")
	cmd.Flags().BoolVar(&ifExists, "if-exists", false, "Don't error if the node doesn't exist")

	return cmd
}
