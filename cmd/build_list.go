package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newBuildListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list <job-path>",
		Short: "List builds for a job",
		Long:  "List recent builds for a Jenkins job.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]

			builds, err := jenkinsClient.ListBuilds(jobPath, limit)
			if err != nil {
				return fmt.Errorf("listing builds: %w", err)
			}

			if len(builds) == 0 {
				fmt.Fprintln(os.Stdout, "No builds found.")
				return nil
			}

			tableDef := &output.TableDef{
				Headers: []string{"NUMBER", "STATUS", "TIMESTAMP", "DURATION"},
				RowFunc: func(item interface{}) []string {
					b := item.(client.Build)
					status := b.Result
					if b.Building {
						status = "RUNNING"
					}
					if status == "" {
						status = "N/A"
					}
					return []string{
						fmt.Sprintf("#%d", b.Number),
						status,
						client.FormatTimestamp(b.Timestamp),
						client.FormatDuration(b.Duration),
					}
				},
			}

			return output.Print(os.Stdout, outFormat, builds, tableDef)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 25, "Maximum number of builds to list")

	return cmd
}
