package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newBuildReplayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "replay <job-path> <build-number>",
		Short: "Replay a pipeline build",
		Long:  "Replay a pipeline build, re-running it with the same pipeline script.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			number, err := client.ParseBuildNumber(args[1])
			if err != nil {
				return err
			}

			if err := jenkinsClient.ReplayBuild(jobPath, number); err != nil {
				return fmt.Errorf("replaying build: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Build #%d replay triggered.\n", number)
			return nil
		},
	}
}
