package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newJobEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable <job-path>",
		Short: "Enable a job",
		Long:  "Enable a disabled Jenkins job.",
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
