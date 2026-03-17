package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newJobBuildCmd() *cobra.Command {
	var params []string
	var wait bool
	var follow bool
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "build <job-path>",
		Short: "Trigger a build",
		Long: `Trigger a build for a Jenkins job. Optionally wait for completion or follow the console log.

Use --param to pass build parameters (repeatable). Use --wait to block
until the build completes. Use --follow to stream the console output in
real time. Use --timeout to set a maximum wait duration (default 30m).

When --wait and --follow are both set, the CLI waits for the build to
complete and then prints the full console output. When only --follow is
set, the CLI streams the console output as the build runs.

Examples:
  # Trigger a simple build
  jenkins job build my-pipeline

  # Trigger a parameterized build
  jenkins job build my-pipeline --param BRANCH=main --param ENV=staging

  # Trigger and wait for completion
  jenkins job build my-pipeline --wait

  # Trigger and stream console output in real time
  jenkins job build my-pipeline --follow

  # Trigger with params, wait, and stream output
  jenkins job build my-pipeline --param BRANCH=main --wait --follow

  # Set a custom timeout (default is 30m)
  jenkins job build my-pipeline --wait --timeout 1h

  # Build a job inside a folder
  jenkins job build my-folder/my-pipeline --param VERSION=1.2.3`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]

			// Parse params
			paramMap := make(map[string]string)
			for _, p := range params {
				parts := strings.SplitN(p, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid parameter format %q, expected KEY=VALUE", p)
				}
				paramMap[parts[0]] = parts[1]
			}

			if wait || follow {
				fmt.Fprintf(os.Stdout, "Triggering build for %q and waiting...\n", jobPath)
				build, err := jenkinsClient.TriggerBuildAndWait(jobPath, paramMap, timeout)
				if err != nil {
					return fmt.Errorf("build failed: %w", err)
				}

				fmt.Fprintf(os.Stdout, "Build #%d completed: %s (duration: %s)\n",
					build.Number, build.Result, client.FormatDuration(build.Duration))

				if follow {
					fmt.Fprintf(os.Stdout, "\n--- Console Output ---\n")
					log, err := jenkinsClient.GetBuildLog(jobPath, build.Number)
					if err != nil {
						return fmt.Errorf("getting build log: %w", err)
					}
					fmt.Fprint(os.Stdout, log)
				}

				return nil
			}

			// Just trigger
			ql, err := jenkinsClient.TriggerBuild(jobPath, paramMap)
			if err != nil {
				return fmt.Errorf("triggering build: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Build triggered for %q.\n", jobPath)
			if ql.QueueURL != "" {
				fmt.Fprintf(os.Stdout, "Queue URL: %s\n", ql.QueueURL)
			}

			if follow {
				// Follow mode without wait: stream log for the new build
				fmt.Fprintf(os.Stdout, "Waiting for build to start...\n")
				buildRef, err := jenkinsClient.WaitForQueuedBuild(ql.QueueURL, timeout)
				if err != nil {
					return fmt.Errorf("waiting for build: %w", err)
				}
				fmt.Fprintf(os.Stdout, "Build #%d started. Streaming console...\n", buildRef.Number)
				return jenkinsClient.StreamBuildLog(jobPath, buildRef.Number, os.Stdout)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&params, "param", "p", nil, "Build parameters (KEY=VALUE, repeatable)")
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "Wait for build to complete")
	cmd.Flags().BoolVarP(&follow, "follow", "F", false, "Follow build console output")
	cmd.Flags().DurationVar(&timeout, "timeout", 30*time.Minute, "Timeout for --wait/--follow")

	return cmd
}
