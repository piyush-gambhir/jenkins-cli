package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newNodeCreateCmd() *cobra.Command {
	var numExecutors int
	var remoteFS string
	var labels string

	cmd := &cobra.Command{
		Use:   "create <node-name>",
		Short: "Create a new node",
		Long:  "Create a new permanent agent node in Jenkins.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if remoteFS == "" {
				return fmt.Errorf("--remote-fs is required")
			}

			if err := jenkinsClient.CreateNode(name, numExecutors, remoteFS, labels); err != nil {
				return fmt.Errorf("creating node: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Node %q created.\n", name)
			return nil
		},
	}

	cmd.Flags().IntVar(&numExecutors, "executors", 1, "Number of executors")
	cmd.Flags().StringVar(&remoteFS, "remote-fs", "", "Remote filesystem root (required)")
	cmd.Flags().StringVar(&labels, "labels", "", "Node labels (space-separated)")

	return cmd
}
