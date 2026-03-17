package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newBuildStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <job-path> <build-number>",
		Short: "Stop a running build",
		Long:  "Stop a currently running build.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			number, err := client.ParseBuildNumber(args[1])
			if err != nil {
				return err
			}

			if err := jenkinsClient.StopBuild(jobPath, number); err != nil {
				return fmt.Errorf("stopping build: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Build #%d stopped.\n", number)
			return nil
		},
	}
}
