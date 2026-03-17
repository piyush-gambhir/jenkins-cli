package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newBuildStagesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stages <job-path> <build-number>",
		Short: "Get pipeline stages",
		Long:  "Display pipeline stage information for a build.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			number, err := parseNumber(args[1])
			if err != nil {
				return err
			}

			run, err := jenkinsClient.GetBuildStages(jobPath, number)
			if err != nil {
				return fmt.Errorf("getting stages: %w", err)
			}

			if len(run.Stages) == 0 {
				fmt.Fprintln(os.Stdout, "No stages found (is this a pipeline job?).")
				return nil
			}

			if outFormat == output.FormatTable {
				fmt.Fprintf(os.Stdout, "Pipeline: %s (Status: %s, Duration: %s)\n\n",
					run.Name, run.Status, client.FormatDuration(run.DurationMillis))

				tableDef := &output.TableDef{
					Headers: []string{"STAGE", "STATUS", "DURATION"},
					RowFunc: func(item interface{}) []string {
						s := item.(client.PipelineStage)
						return []string{
							s.Name,
							s.Status,
							client.FormatDuration(s.DurationMillis),
						}
					},
				}
				return output.Print(os.Stdout, outFormat, run.Stages, tableDef)
			}

			return output.Print(os.Stdout, outFormat, run, nil)
		},
	}
}
