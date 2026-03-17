package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newQueueListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List queued items",
		Long: `List all items currently in the Jenkins build queue.

Shows each queued item's ID, task name, reason for being queued, and
whether it is stuck or blocked. An empty queue means all builds have
been assigned executors.

Examples:
  # List all queued builds
  jenkins queue list

  # Output as JSON
  jenkins queue list -o json`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := jenkinsClient.ListQueue()
			if err != nil {
				return fmt.Errorf("listing queue: %w", err)
			}

			if len(items) == 0 {
				fmt.Fprintln(os.Stdout, "Build queue is empty.")
				return nil
			}

			tableDef := &output.TableDef{
				Headers: []string{"ID", "TASK", "WHY", "STUCK", "BLOCKED"},
				RowFunc: func(item interface{}) []string {
					q := item.(client.QueueItem)
					why := q.Why
					if len(why) > 60 {
						why = why[:60] + "..."
					}
					return []string{
						fmt.Sprintf("%d", q.ID),
						q.Task.Name,
						why,
						fmt.Sprintf("%v", q.Stuck),
						fmt.Sprintf("%v", q.Blocked),
					}
				},
			}

			return output.Print(os.Stdout, outFormat, items, tableDef)
		},
	}
}
