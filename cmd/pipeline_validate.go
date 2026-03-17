package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newPipelineValidateCmd() *cobra.Command {
	var fromFile string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a Jenkinsfile",
		Long:  "Validate a declarative Jenkinsfile using the Jenkins pipeline model converter.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if fromFile == "" {
				return fmt.Errorf("--from-file is required")
			}

			data, err := os.ReadFile(fromFile)
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

	cmd.Flags().StringVarP(&fromFile, "from-file", "f", "", "Path to Jenkinsfile (required)")

	return cmd
}
