package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newJobDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "disable <job-path>",
		Short:       "Disable a job",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Disable a Jenkins job so it cannot be built.

A disabled job cannot be triggered manually or by SCM changes. Use
"jenkins job enable" to re-enable it.

Examples:
  # Disable a job
  jenkins job disable my-pipeline

  # Disable a job in a folder
  jenkins job disable my-folder/my-pipeline`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]

			if err := jenkinsClient.DisableJob(jobPath); err != nil {
				return fmt.Errorf("disabling job: %w", err)
			}

			if !quietFlag {
				fmt.Fprintf(os.Stdout, "Job %q disabled.\n", jobPath)
			}
			return nil
		},
	}
}
