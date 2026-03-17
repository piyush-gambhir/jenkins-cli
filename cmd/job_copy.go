package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newJobCopyCmd() *cobra.Command {
	var folder string

	cmd := &cobra.Command{
		Use:   "copy <source-job> <new-name>",
		Short: "Copy a job",
		Long: `Create a copy of an existing Jenkins job with a new name.

Copies the source job's entire configuration to a new job. Use --folder
to specify the folder context for the copy operation.

Examples:
  # Copy a job at the root level
  jenkins job copy my-pipeline my-pipeline-copy

  # Copy a job within a folder
  jenkins job copy my-pipeline new-pipeline --folder my-folder`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]
			newName := args[1]

			if err := jenkinsClient.CopyJob(source, newName, folder); err != nil {
				return fmt.Errorf("copying job: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Job %q copied to %q successfully.\n", source, newName)
			return nil
		},
	}

	cmd.Flags().StringVarP(&folder, "folder", "f", "", "Folder context for copy operation")

	return cmd
}
