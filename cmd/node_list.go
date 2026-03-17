package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newNodeListCmd() *cobra.Command {
	var offlineOnly bool
	var onlineOnly bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List nodes",
		Long: `List all Jenkins nodes/agents.

Displays each node's name, executor count, idle status, offline status,
and offline reason (if any). Use --offline or --online to filter by
connectivity state. These two flags are mutually exclusive.

Examples:
  # List all nodes
  jenkins node list

  # List only offline nodes
  jenkins node list --offline

  # List only online nodes
  jenkins node list --online

  # Output as JSON
  jenkins node list -o json

  # Output as YAML
  jenkins node list -o yaml`,
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if offlineOnly && onlineOnly {
				return fmt.Errorf("--offline and --online are mutually exclusive")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			nodes, err := jenkinsClient.ListNodes()
			if err != nil {
				return fmt.Errorf("listing nodes: %w", err)
			}

			// Filter by online/offline status
			if offlineOnly {
				var filtered []client.Node
				for _, n := range nodes {
					if n.Offline {
						filtered = append(filtered, n)
					}
				}
				nodes = filtered
			} else if onlineOnly {
				var filtered []client.Node
				for _, n := range nodes {
					if !n.Offline {
						filtered = append(filtered, n)
					}
				}
				nodes = filtered
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

	cmd.Flags().BoolVar(&offlineOnly, "offline", false, "Show only offline nodes")
	cmd.Flags().BoolVar(&onlineOnly, "online", false, "Show only online nodes")

	return cmd
}
