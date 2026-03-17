package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newJobDeleteCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <job-path>",
		Short: "Delete a job",
		Long:  "Permanently delete a Jenkins job. Use --confirm to skip the confirmation prompt.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]

			if !confirm {
				return fmt.Errorf("use --confirm to confirm deletion of job %q", jobPath)
			}

			if err := jenkinsClient.DeleteJob(jobPath); err != nil {
				return fmt.Errorf("deleting job: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Job %q deleted successfully.\n", jobPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm deletion")

	return cmd
}
