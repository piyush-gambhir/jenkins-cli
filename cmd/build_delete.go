package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newBuildDeleteCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <job-path> <build-number>",
		Short: "Delete a build",
		Long: `Permanently delete a build record.

WARNING: This operation is irreversible. The build record, console log,
and artifacts will be permanently removed. Requires --confirm.

Examples:
  # Delete build #42
  jenkins build delete my-pipeline 42 --confirm

  # Delete a build for a job in a folder
  jenkins build delete my-folder/my-pipeline 10 --confirm`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			number, err := client.ParseBuildNumber(args[1])
			if err != nil {
				return err
			}

			if !confirm {
				return fmt.Errorf("use --confirm to confirm deletion of build #%d", number)
			}

			if err := jenkinsClient.DeleteBuild(jobPath, number); err != nil {
				return fmt.Errorf("deleting build: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Build #%d deleted.\n", number)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm deletion")

	return cmd
}
