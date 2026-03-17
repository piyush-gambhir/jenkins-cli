package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newPipelineInputListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "input-list <job-path> <build-number>",
		Short: "List pending pipeline inputs",
		Long: `List pending input actions for a pipeline build.

When a pipeline has an "input" step, the build pauses and waits for
user action. This command lists all such pending inputs including their
ID, message, and available actions (proceed/abort text).

Examples:
  # List pending inputs for build #42
  jenkins pipeline input-list my-pipeline 42

  # Output as JSON
  jenkins pipeline input-list my-pipeline 42 -o json`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			number, err := parseNumber(args[1])
			if err != nil {
				return err
			}

			inputs, err := jenkinsClient.ListPipelineInputs(jobPath, number)
			if err != nil {
				return fmt.Errorf("listing pipeline inputs: %w", err)
			}

			if len(inputs) == 0 {
				fmt.Fprintln(os.Stdout, "No pending inputs found.")
				return nil
			}

			tableDef := &output.TableDef{
				Headers: []string{"ID", "MESSAGE", "PROCEED", "ABORT"},
				RowFunc: func(item interface{}) []string {
					i := item.(client.PipelineInput)
					msg := i.Message
					if len(msg) > 50 {
						msg = msg[:50] + "..."
					}
					return []string{i.ID, msg, i.ProceedText, i.AbortText}
				},
			}

			return output.Print(os.Stdout, outFormat, inputs, tableDef)
		},
	}
}
