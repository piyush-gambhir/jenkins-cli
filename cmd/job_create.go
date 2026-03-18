package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newJobCreateCmd() *cobra.Command {
	var fromFile string
	var folder string

	cmd := &cobra.Command{
		Use:         "create <job-name>",
		Short:       "Create a new job",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Create a new Jenkins job from an XML configuration file.

The --from-file flag is required and must point to a valid Jenkins
config.xml file. Use --folder to create the job inside a specific folder.

Examples:
  # Create a job at the root level
  jenkins job create my-new-job --from-file config.xml

  # Create a job in a folder
  jenkins job create my-new-job --from-file config.xml --folder my-folder

  # Create a job in a nested folder
  jenkins job create deploy --from-file pipeline-config.xml --folder team/project`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if fromFile == "" {
				return fmt.Errorf("--from-file is required")
			}

			data, err := os.ReadFile(fromFile)
			if err != nil {
				return fmt.Errorf("reading config file %s: %w", fromFile, err)
			}

			if err := jenkinsClient.CreateJob(folder, name, string(data)); err != nil {
				return fmt.Errorf("creating job: %w", err)
			}

			loc := name
			if folder != "" {
				loc = filepath.Join(folder, name)
			}
			fmt.Fprintf(os.Stdout, "Job %q created successfully.\n", loc)
			return nil
		},
	}

	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to XML config file (required)")
	cmd.Flags().StringVarP(&folder, "folder", "f", "", "Folder to create the job in")

	return cmd
}
