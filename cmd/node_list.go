package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newNodeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List nodes",
		Long:  "List all Jenkins nodes/agents.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			nodes, err := jenkinsClient.ListNodes()
			if err != nil {
				return fmt.Errorf("listing nodes: %w", err)
			}

			if len(nodes) == 0 {
				fmt.Fprintln(os.Stdout, "No nodes found.")
				return nil
			}

			tableDef := &output.TableDef{
				Headers: []string{"NAME", "EXECUTORS", "IDLE", "OFFLINE", "REASON"},
				RowFunc: func(item interface{}) []string {
					n := item.(client.Node)
					reason := ""
					if n.OfflineCauseReason != "" {
						reason = n.OfflineCauseReason
						if len(reason) > 50 {
							reason = reason[:50] + "..."
						}
					}
					return []string{
						n.DisplayName,
						fmt.Sprintf("%d", n.NumExecutors),
						fmt.Sprintf("%v", n.Idle),
						fmt.Sprintf("%v", n.Offline),
						reason,
					}
				},
			}

			return output.Print(os.Stdout, outFormat, nodes, tableDef)
		},
	}
}
