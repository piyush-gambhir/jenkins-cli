package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newNodeGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <node-name>",
		Short: "Get node details",
		Long:  "Display detailed information about a specific node.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			node, err := jenkinsClient.GetNode(name)
			if err != nil {
				return fmt.Errorf("getting node: %w", err)
			}

			if outFormat == output.FormatTable {
				fmt.Fprintf(os.Stdout, "Node: %s\n", node.DisplayName)
				fmt.Fprintf(os.Stdout, "  Description:  %s\n", node.Description)
				fmt.Fprintf(os.Stdout, "  Executors:    %d\n", node.NumExecutors)
				fmt.Fprintf(os.Stdout, "  Idle:         %v\n", node.Idle)
				fmt.Fprintf(os.Stdout, "  Offline:      %v\n", node.Offline)
				fmt.Fprintf(os.Stdout, "  Temp Offline: %v\n", node.TemporarilyOffline)
				fmt.Fprintf(os.Stdout, "  JNLP Agent:   %v\n", node.JNLPAgent)
				if node.OfflineCauseReason != "" {
					fmt.Fprintf(os.Stdout, "  Offline Reason: %s\n", node.OfflineCauseReason)
				}
				return nil
			}

			return output.Print(os.Stdout, outFormat, node, nil)
		},
	}
}
