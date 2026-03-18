package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newNodeDeleteCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:         "delete <node-name>",
		Short:       "Delete a node",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Permanently delete a Jenkins node/agent.

WARNING: This operation is irreversible. Requires --confirm.

Examples:
  # Delete a node
  jenkins node delete my-agent --confirm`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !confirm {
				return fmt.Errorf("use --confirm to confirm deletion of node %q", name)
			}

			if err := jenkinsClient.DeleteNode(name); err != nil {
				return fmt.Errorf("deleting node: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Node %q deleted.\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm deletion")

	return cmd
}
