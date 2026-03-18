package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newJobDeleteCmd() *cobra.Command {
	var confirm bool
	var ifExists bool

	cmd := &cobra.Command{
		Use:         "delete <job-path>",
		Short:       "Delete a job",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Permanently delete a Jenkins job. Use --confirm to skip the confirmation prompt.

WARNING: This operation is irreversible. The job and all its build
history will be permanently removed.

Examples:
  # Delete a job (requires --confirm)
  jenkins job delete my-pipeline --confirm

  # Delete a job in a folder
  jenkins job delete my-folder/my-pipeline --confirm

  # Idempotent delete (no error if job doesn't exist)
  jenkins job delete my-pipeline --confirm --if-exists`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]

			if !confirm {
				if noInputFlag {
					return fmt.Errorf("interactive input required but --no-input is set. Use --confirm for destructive operations.")
				}
				return fmt.Errorf("use --confirm to confirm deletion of job %q", jobPath)
			}

			if err := jenkinsClient.DeleteJob(jobPath); err != nil {
				var apiErr *client.APIError
				if ifExists && errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
					if !quietFlag {
						fmt.Fprintf(os.Stdout, "Job %q does not exist, skipping.\n", jobPath)
					}
					return nil
				}
				return fmt.Errorf("deleting job: %w", err)
			}

			if !quietFlag {
				fmt.Fprintf(os.Stdout, "Job %q deleted successfully.\n", jobPath)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm deletion")
	cmd.Flags().BoolVar(&ifExists, "if-exists", false, "Don't error if the job doesn't exist")

	return cmd
}
