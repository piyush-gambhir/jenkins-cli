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
		Long: `Replay a pipeline build, re-running it with the same pipeline script.

Triggers a new build using the same Jenkinsfile/pipeline script that was
used in the specified build. This is useful for re-running a build without
committing changes to SCM.

Examples:
  # Replay build #42
  jenkins build replay my-pipeline 42

  # Replay a build for a job in a folder
  jenkins build replay my-folder/my-pipeline 10`,
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
