package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newJobListCmd() *cobra.Command {
	var folder string
	var recursive bool
	var status string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Jenkins jobs",
		Long: `List jobs in the root or a specific folder.

By default lists jobs at the root level. Use --folder to list jobs in a
specific folder, and --recursive to traverse all subfolders. Use --status
to filter jobs by their current status (derived from the job color).

Valid status values: SUCCESS, FAILURE, UNSTABLE, DISABLED, ABORTED, NOT_BUILT, RUNNING.

Examples:
  # List all root-level jobs
  jenkins job list

  # List jobs in a folder
  jenkins job list --folder my-folder

  # List all jobs recursively (all folders)
  jenkins job list --recursive

  # List jobs in a folder recursively
  jenkins job list --folder my-team --recursive

  # List only failed jobs
  jenkins job list --status FAILURE

  # List only disabled jobs
  jenkins job list --status DISABLED

  # List jobs recursively and filter by status
  jenkins job list --recursive --status SUCCESS

  # Output as JSON
  jenkins job list -o json

  # Output as YAML
  jenkins job list -o yaml`,
		Args: cobra.NoArgs,
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

			// Filter by status if specified
			if status != "" {
				statusUpper := strings.ToUpper(status)
				var filtered []client.Job
				for _, j := range jobs {
					jobStatus := client.ColorToStatus(j.Color)
					if strings.EqualFold(jobStatus, statusUpper) {
						filtered = append(filtered, j)
					}
				}
				jobs = filtered
			}

			if len(jobs) == 0 {
				fmt.Fprintln(os.Stdout, "No jobs found.")
				return nil
			}

			tableDef := &output.TableDef{
				Headers: []string{"NAME", "STATUS", "LAST BUILD", "RESULT"},
				RowFunc: func(item interface{}) []string {
					j := item.(client.Job)
					jobStatus := client.ColorToStatus(j.Color)
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
					return []string{name, jobStatus, lastBuild, result}
				},
			}

			return output.Print(os.Stdout, outFormat, jobs, tableDef)
		},
	}

	cmd.Flags().StringVarP(&folder, "folder", "f", "", "Folder path to list jobs from")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "List jobs recursively through all subfolders")
	cmd.Flags().StringVar(&status, "status", "", "Filter by job status (SUCCESS, FAILURE, UNSTABLE, DISABLED, ABORTED, NOT_BUILT, RUNNING)")

	return cmd
}
