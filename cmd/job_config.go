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
		Long:  "Retrieve and display the config.xml of a Jenkins job.",
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
