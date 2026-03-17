package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newJobUpdateCmd() *cobra.Command {
	var fromFile string

	cmd := &cobra.Command{
		Use:   "update <job-path>",
		Short: "Update job configuration",
		Long: `Update a Jenkins job's config.xml with a new XML configuration.

The --from-file flag is required. This replaces the entire config.xml
of the specified job. To view the current config first, use
"jenkins job config <job-path>".

Examples:
  # Update a job's configuration
  jenkins job update my-pipeline --from-file new-config.xml

  # Update a job in a folder
  jenkins job update my-folder/my-pipeline --from-file config.xml

  # Export, edit, and re-import a job config
  jenkins job config my-pipeline > config.xml
  # ... edit config.xml ...
  jenkins job update my-pipeline --from-file config.xml`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]

			if fromFile == "" {
				return fmt.Errorf("--from-file is required")
			}

			data, err := os.ReadFile(fromFile)
			if err != nil {
				return fmt.Errorf("reading config file %s: %w", fromFile, err)
			}

			if err := jenkinsClient.UpdateJobConfig(jobPath, string(data)); err != nil {
				return fmt.Errorf("updating job: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Job %q updated successfully.\n", jobPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to XML config file (required)")

	return cmd
}
