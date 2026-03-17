package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newNodeLogCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "log <node-name>",
		Short: "Get node agent log",
		Long: `Display the agent log for a Jenkins node.

Shows the agent connection log, which is useful for diagnosing
connectivity issues between the controller and the agent.

Examples:
  # View agent log for a node
  jenkins node log my-agent

  # Pipe log to a file
  jenkins node log my-agent > agent.log`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			log, err := jenkinsClient.GetNodeLog(name)
			if err != nil {
				return fmt.Errorf("getting node log: %w", err)
			}

			fmt.Fprint(os.Stdout, log)
			return nil
		},
	}
}
