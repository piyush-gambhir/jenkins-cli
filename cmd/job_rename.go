package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newJobRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <job-path> <new-name>",
		Short: "Rename a job",
		Long:  "Rename an existing Jenkins job.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			newName := args[1]

			if err := jenkinsClient.RenameJob(jobPath, newName); err != nil {
				return fmt.Errorf("renaming job: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Job %q renamed to %q successfully.\n", jobPath, newName)
			return nil
		},
	}
}
