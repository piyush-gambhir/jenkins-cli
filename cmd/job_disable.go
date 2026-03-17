package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newJobDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <job-path>",
		Short: "Disable a job",
		Long:  "Disable a Jenkins job so it cannot be built.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]

			if err := jenkinsClient.DisableJob(jobPath); err != nil {
				return fmt.Errorf("disabling job: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Job %q disabled.\n", jobPath)
			return nil
		},
	}
}
