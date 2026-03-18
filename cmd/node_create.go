package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newNodeCreateCmd() *cobra.Command {
	var numExecutors int
	var remoteFS string
	var labels string
	var ifNotExists bool

	cmd := &cobra.Command{
		Use:         "create <node-name>",
		Short:       "Create a new node",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Create a new permanent agent node in Jenkins.

Creates a JNLP (Java Web Start) permanent agent. The --remote-fs flag
is required and specifies the remote filesystem root directory on the
agent machine.

Examples:
  # Create a node with 2 executors
  jenkins node create my-agent --remote-fs /home/jenkins --executors 2

  # Create a node with labels
  jenkins node create build-agent --remote-fs /opt/jenkins --labels "linux docker"

  # Idempotent create (no error if node already exists)
  jenkins node create my-agent --remote-fs /var/jenkins --if-not-exists`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if remoteFS == "" {
				return fmt.Errorf("--remote-fs is required")
			}

			if err := jenkinsClient.CreateNode(name, numExecutors, remoteFS, labels); err != nil {
				var apiErr *client.APIError
				if ifNotExists && errors.As(err, &apiErr) && (apiErr.StatusCode == 400 || apiErr.StatusCode == 409) {
					if !quietFlag {
						fmt.Fprintf(os.Stdout, "Node %q already exists, skipping.\n", name)
					}
					return nil
				}
				return fmt.Errorf("creating node: %w", err)
			}

			if !quietFlag {
				fmt.Fprintf(os.Stdout, "Node %q created.\n", name)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&numExecutors, "executors", 1, "Number of executors")
	cmd.Flags().StringVar(&remoteFS, "remote-fs", "", "Remote filesystem root (required)")
	cmd.Flags().StringVar(&labels, "labels", "", "Node labels (space-separated)")
	cmd.Flags().BoolVar(&ifNotExists, "if-not-exists", false, "Don't error if the node already exists")

	return cmd
}
