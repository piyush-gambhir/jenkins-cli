package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newBuildLogCmd() *cobra.Command {
	var follow bool

	cmd := &cobra.Command{
		Use:   "log <job-path> <build-number>",
		Short: "Get build console output",
		Long: `Display the console output of a build. Use --follow to stream output in real-time.

Without --follow, the entire console log is fetched and printed at once.
With --follow, the log is streamed progressively as the build runs,
similar to "tail -f". This is particularly useful for monitoring builds
in progress.

Examples:
  # View the full console log of build #42
  jenkins build log my-pipeline 42

  # Stream the log of a running build in real-time
  jenkins build log my-pipeline 42 --follow

  # View log for a job in a folder
  jenkins build log my-folder/my-pipeline 10

  # Pipe log output to a file
  jenkins build log my-pipeline 42 > build-42.log`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			number, err := client.ParseBuildNumber(args[1])
			if err != nil {
				return err
			}

			if follow {
				return jenkinsClient.StreamBuildLog(jobPath, number, os.Stdout)
			}

			log, err := jenkinsClient.GetBuildLog(jobPath, number)
			if err != nil {
				return fmt.Errorf("getting build log: %w", err)
			}

			fmt.Fprint(os.Stdout, log)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow (stream) the log output")

	return cmd
}
