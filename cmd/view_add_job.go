package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newViewAddJobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-job <view-name> <job-name>",
		Short: "Add a job to a view",
		Long:  "Add an existing job to a Jenkins view.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			viewName := args[0]
			jobName := args[1]

			if err := jenkinsClient.AddJobToView(viewName, jobName); err != nil {
				return fmt.Errorf("adding job to view: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Job %q added to view %q.\n", jobName, viewName)
			return nil
		},
	}
}
