package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func newPipelineValidateCmd() *cobra.Command {
	var fromFile string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a Jenkinsfile",
		Long: `Validate a declarative Jenkinsfile using the Jenkins pipeline model converter.

Sends the Jenkinsfile content to the Jenkins server for validation.
Returns any syntax errors or "Jenkinsfile successfully validated" on
success. The --from-file flag is required. Use "-" to read from stdin.

Note: Only declarative pipelines can be validated. Scripted pipelines
are not supported by this endpoint.

Examples:
  # Validate a Jenkinsfile
  jenkins pipeline validate --from-file Jenkinsfile

  # Validate from stdin
  cat Jenkinsfile | jenkins pipeline validate --from-file -`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if fromFile == "" {
				return fmt.Errorf("--from-file is required")
			}

			var data []byte
			var err error
			if fromFile == "-" {
				data, err = io.ReadAll(os.Stdin)
			} else {
				data, err = os.ReadFile(fromFile)
			}
			if err != nil {
				return fmt.Errorf("reading file %s: %w", fromFile, err)
			}

			result, err := jenkinsClient.ValidateJenkinsfile(string(data))
			if err != nil {
				return fmt.Errorf("validating Jenkinsfile: %w", err)
			}

			fmt.Fprint(os.Stdout, result)
			return nil
		},
	}

	cmd.Flags().StringVarP(&fromFile, "from-file", "f", "", "Path to Jenkinsfile (required, use - for stdin)")

	return cmd
}
