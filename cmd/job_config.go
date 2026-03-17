package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newJobConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config <job-path>",
		Short: "Get job configuration XML",
		Long: `Retrieve and display the config.xml of a Jenkins job.

Outputs the raw XML configuration. This can be saved to a file, edited,
and re-imported using "jenkins job update".

Examples:
  # Print a job's config.xml to stdout
  jenkins job config my-pipeline

  # Save config to a file
  jenkins job config my-pipeline > config.xml

  # View config for a job in a folder
  jenkins job config my-folder/my-pipeline`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]

			configXML, err := jenkinsClient.GetJobConfig(jobPath)
			if err != nil {
				return fmt.Errorf("getting job config: %w", err)
			}

			fmt.Fprint(os.Stdout, configXML)
			return nil
		},
	}
}
