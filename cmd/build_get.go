package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newBuildGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <job-path> <build-number>",
		Short: "Get build details",
		Long: `Display detailed information about a specific build.

Shows build number, display name, URL, result, building status, timestamp,
duration, description, artifacts, and change set information.

Examples:
  # Get details about build #42
  jenkins build get my-pipeline 42

  # Get build info for a job in a folder
  jenkins build get my-folder/my-pipeline 10

  # Output as JSON
  jenkins build get my-pipeline 42 -o json

  # Output as YAML
  jenkins build get my-pipeline 42 -o yaml`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			number, err := client.ParseBuildNumber(args[1])
			if err != nil {
				return err
			}

			build, err := jenkinsClient.GetBuild(jobPath, number)
			if err != nil {
				return fmt.Errorf("getting build: %w", err)
			}

			if outFormat == output.FormatTable {
				fmt.Fprintf(os.Stdout, "Build #%d\n", build.Number)
				fmt.Fprintf(os.Stdout, "  Display Name: %s\n", build.DisplayName)
				fmt.Fprintf(os.Stdout, "  URL:          %s\n", build.URL)
				fmt.Fprintf(os.Stdout, "  Result:       %s\n", build.Result)
				fmt.Fprintf(os.Stdout, "  Building:     %v\n", build.Building)
				fmt.Fprintf(os.Stdout, "  Timestamp:    %s\n", client.FormatTimestamp(build.Timestamp))
				fmt.Fprintf(os.Stdout, "  Duration:     %s\n", client.FormatDuration(build.Duration))
				if build.Description != "" {
					fmt.Fprintf(os.Stdout, "  Description:  %s\n", build.Description)
				}
				if len(build.Artifacts) > 0 {
					fmt.Fprintf(os.Stdout, "  Artifacts:    %d\n", len(build.Artifacts))
					for _, a := range build.Artifacts {
						fmt.Fprintf(os.Stdout, "    - %s\n", a.FileName)
					}
				}
				if build.ChangeSet != nil && len(build.ChangeSet.Items) > 0 {
					fmt.Fprintf(os.Stdout, "  Changes:      %d\n", len(build.ChangeSet.Items))
					for _, c := range build.ChangeSet.Items {
						fmt.Fprintf(os.Stdout, "    - %s (%s)\n", c.Message, c.Author.FullName)
					}
				}
				return nil
			}

			return output.Print(os.Stdout, outFormat, build, nil)
		},
	}
}
