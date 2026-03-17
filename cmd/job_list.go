package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newJobListCmd() *cobra.Command {
	var folder string
	var recursive bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Jenkins jobs",
		Long:  "List jobs in the root or a specific folder. Use --recursive to list all jobs recursively.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var jobs []client.Job
			var err error

			if recursive {
				jobs, err = jenkinsClient.ListJobsRecursive(folder)
			} else {
				jobs, err = jenkinsClient.ListJobs(folder)
			}
			if err != nil {
				return fmt.Errorf("listing jobs: %w", err)
			}

			if len(jobs) == 0 {
				fmt.Fprintln(os.Stdout, "No jobs found.")
				return nil
			}

			tableDef := &output.TableDef{
				Headers: []string{"NAME", "STATUS", "LAST BUILD", "RESULT"},
				RowFunc: func(item interface{}) []string {
					j := item.(client.Job)
					status := client.ColorToStatus(j.Color)
					lastBuild := "N/A"
					result := "N/A"
					if j.LastBuild != nil {
						lastBuild = fmt.Sprintf("#%d", j.LastBuild.Number)
						if j.LastBuild.Result != "" {
							result = j.LastBuild.Result
						} else {
							result = "RUNNING"
						}
					}
					name := j.Name
					if j.FullName != "" {
						name = j.FullName
					}
					return []string{name, status, lastBuild, result}
				},
			}

			return output.Print(os.Stdout, outFormat, jobs, tableDef)
		},
	}

	cmd.Flags().StringVarP(&folder, "folder", "f", "", "Folder path to list jobs from")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "List jobs recursively")

	return cmd
}
