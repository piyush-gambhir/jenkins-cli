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
		Use:   "run-script",
		Short: "Execute a Groovy script",
		Long:  "Execute a Groovy script on the Jenkins controller via the script console.",
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
