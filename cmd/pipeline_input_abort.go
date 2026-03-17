package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newPipelineInputAbortCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "input-abort <job-path> <build-number> <input-id>",
		Short: "Abort a pipeline input",
		Long: `Abort a pending pipeline input action.

Rejects the input step, which causes the pipeline to proceed down the
abort path (or fail, depending on how the Jenkinsfile handles it).

Examples:
  # Abort a pending input
  jenkins pipeline input-abort my-pipeline 42 my-input-id`,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			number, err := parseNumber(args[1])
			if err != nil {
				return err
			}
			inputID := args[2]

			if err := jenkinsClient.AbortPipelineInput(jobPath, number, inputID); err != nil {
				return fmt.Errorf("aborting input: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Input %q aborted for build #%d.\n", inputID, number)
			return nil
		},
	}
}
