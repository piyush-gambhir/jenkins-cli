package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newJobRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "rename <job-path> <new-name>",
		Short:       "Rename a job",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Rename an existing Jenkins job.

This changes the job's name while keeping all other configuration,
build history, and workspace intact.

Examples:
  # Rename a root-level job
  jenkins job rename old-name new-name

  # Rename a job in a folder
  jenkins job rename my-folder/old-name new-name`,
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
