package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newViewRemoveJobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove-job <view-name> <job-name>",
		Short: "Remove a job from a view",
		Long:  "Remove a job from a Jenkins view.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			viewName := args[0]
			jobName := args[1]

			if err := jenkinsClient.RemoveJobFromView(viewName, jobName); err != nil {
				return fmt.Errorf("removing job from view: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Job %q removed from view %q.\n", jobName, viewName)
			return nil
		},
	}
}
