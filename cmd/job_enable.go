package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newJobEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "enable <job-path>",
		Short:       "Enable a job",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Enable a disabled Jenkins job.

Re-enables a job that was previously disabled so it can be built again.

Examples:
  # Enable a job
  jenkins job enable my-pipeline

  # Enable a job in a folder
  jenkins job enable my-folder/my-pipeline`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]

			if err := jenkinsClient.EnableJob(jobPath); err != nil {
				return fmt.Errorf("enabling job: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Job %q enabled.\n", jobPath)
			return nil
		},
	}
}
