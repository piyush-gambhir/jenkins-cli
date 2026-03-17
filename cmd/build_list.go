package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newBuildListCmd() *cobra.Command {
	var limit int
	var status string

	cmd := &cobra.Command{
		Use:   "list <job-path>",
		Short: "List builds for a job",
		Long: `List recent builds for a Jenkins job.

Returns build number, status, timestamp, and duration for each build.
Use --status to filter by build result. Use --limit to control how many
builds are returned (default 25).

Valid status values: SUCCESS, FAILURE, UNSTABLE, ABORTED, NOT_BUILT, RUNNING.

Examples:
  # List the last 25 builds of a job
  jenkins build list my-pipeline

  # List the last 10 builds
  jenkins build list my-pipeline --limit 10

  # List only failed builds
  jenkins build list my-pipeline --status FAILURE

  # List only successful builds (max 50)
  jenkins build list my-pipeline --status SUCCESS --limit 50

  # List builds for a job inside a folder
  jenkins build list my-folder/my-pipeline

  # Output as JSON for programmatic use
  jenkins build list my-pipeline -o json

  # Output as YAML
  jenkins build list my-pipeline -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]

			builds, err := jenkinsClient.ListBuilds(jobPath, limit)
			if err != nil {
				return fmt.Errorf("listing builds: %w", err)
			}

			// Filter by status if specified
			if status != "" {
				statusUpper := strings.ToUpper(status)
				var filtered []client.Build
				for _, b := range builds {
					buildStatus := b.Result
					if b.Building {
						buildStatus = "RUNNING"
					}
					if strings.EqualFold(buildStatus, statusUpper) {
						filtered = append(filtered, b)
					}
				}
				builds = filtered
			}

			if len(builds) == 0 {
				fmt.Fprintln(os.Stdout, "No builds found.")
				return nil
			}

			tableDef := &output.TableDef{
				Headers: []string{"NUMBER", "STATUS", "TIMESTAMP", "DURATION"},
				RowFunc: func(item interface{}) []string {
					b := item.(client.Build)
					s := b.Result
					if b.Building {
						s = "RUNNING"
					}
					if s == "" {
						s = "N/A"
					}
					return []string{
						fmt.Sprintf("#%d", b.Number),
						s,
						client.FormatTimestamp(b.Timestamp),
						client.FormatDuration(b.Duration),
					}
				},
			}

			return output.Print(os.Stdout, outFormat, builds, tableDef)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 25, "Maximum number of builds to list")
	cmd.Flags().StringVar(&status, "status", "", "Filter by build status (SUCCESS, FAILURE, UNSTABLE, ABORTED, NOT_BUILT, RUNNING)")

	return cmd
}
