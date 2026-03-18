package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newSystemRunScriptCmd() *cobra.Command {
	var fromFile string
	var script string

	cmd := &cobra.Command{
		Use:         "run-script",
		Short:       "Execute a Groovy script",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Execute a Groovy script on the Jenkins controller via the script console.

Runs an arbitrary Groovy script on the Jenkins controller and prints
the output. Provide the script inline via --script or from a file via
--from-file. One of the two is required.

WARNING: Scripts run with full Jenkins controller access. Use with caution.

Examples:
  # Run an inline Groovy script
  jenkins system run-script --script 'println Jenkins.instance.numExecutors'

  # Run a script from a file
  jenkins system run-script --from-file my-script.groovy

  # Get system properties
  jenkins system run-script --script 'System.getProperties().each { k, v -> println "$k=$v" }'

  # List all jobs via Groovy
  jenkins system run-script --script 'Jenkins.instance.allItems.each { println it.fullName }'`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var scriptContent string

			if fromFile != "" {
				data, err := os.ReadFile(fromFile)
				if err != nil {
					return fmt.Errorf("reading script file %s: %w", fromFile, err)
				}
				scriptContent = string(data)
			} else if script != "" {
				scriptContent = script
			} else {
				return fmt.Errorf("either --script or --from-file is required")
			}

			result, err := jenkinsClient.RunScript(scriptContent)
			if err != nil {
				return fmt.Errorf("running script: %w", err)
			}

			fmt.Fprint(os.Stdout, result)
			return nil
		},
	}

	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to Groovy script file")
	cmd.Flags().StringVar(&script, "script", "", "Groovy script to execute")

	return cmd
}
