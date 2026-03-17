package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newBuildEnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "env <job-path> <build-number>",
		Short: "Get build environment variables",
		Long: `Display the injected environment variables for a build.

Lists all environment variables that were injected into the build
environment, sorted alphabetically by key. Requires the Environment
Injector plugin to be installed on Jenkins.

Examples:
  # View env vars for build #42
  jenkins build env my-pipeline 42

  # Output as JSON for scripting
  jenkins build env my-pipeline 42 -o json

  # Output as YAML
  jenkins build env my-pipeline 42 -o yaml`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			number, err := parseNumber(args[1])
			if err != nil {
				return err
			}

			envVars, err := jenkinsClient.GetBuildEnvVars(jobPath, number)
			if err != nil {
				return fmt.Errorf("getting env vars: %w", err)
			}

			if len(envVars) == 0 {
				fmt.Fprintln(os.Stdout, "No environment variables found.")
				return nil
			}

			if outFormat == output.FormatTable {
				// Sort keys for consistent output
				keys := make([]string, 0, len(envVars))
				for k := range envVars {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				for _, k := range keys {
					fmt.Fprintf(os.Stdout, "%s=%s\n", k, envVars[k])
				}
				return nil
			}

			return output.Print(os.Stdout, outFormat, envVars, nil)
		},
	}
}

func parseNumber(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", s)
	}
	return n, nil
}
