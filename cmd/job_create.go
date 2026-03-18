package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newJobCreateCmd() *cobra.Command {
	var fromFile string
	var folder string
	var ifNotExists bool

	cmd := &cobra.Command{
		Use:         "create <job-name>",
		Short:       "Create a new job",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Create a new Jenkins job from an XML configuration file.

The --from-file flag is required and must point to a valid Jenkins
config.xml file. Use --folder to create the job inside a specific folder.
Use "-" as the file path to read from stdin.

Examples:
  # Create a job at the root level
  jenkins job create my-new-job --from-file config.xml

  # Create a job in a folder
  jenkins job create my-new-job --from-file config.xml --folder my-folder

  # Create a job from stdin
  cat config.xml | jenkins job create my-new-job --from-file -

  # Idempotent create (no error if job already exists)
  jenkins job create my-new-job --from-file config.xml --if-not-exists`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

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
				return fmt.Errorf("reading config file %s: %w", fromFile, err)
			}

			if err := jenkinsClient.CreateJob(folder, name, string(data)); err != nil {
				var apiErr *client.APIError
				if ifNotExists && errors.As(err, &apiErr) && (apiErr.StatusCode == 400 || apiErr.StatusCode == 409) {
					loc := name
					if folder != "" {
						loc = filepath.Join(folder, name)
					}
					if !quietFlag {
						fmt.Fprintf(os.Stdout, "Job %q already exists, skipping.\n", loc)
					}
					return nil
				}
				return fmt.Errorf("creating job: %w", err)
			}

			loc := name
			if folder != "" {
				loc = filepath.Join(folder, name)
			}
			if !quietFlag {
				fmt.Fprintf(os.Stdout, "Job %q created successfully.\n", loc)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to XML config file (required, use - for stdin)")
	cmd.Flags().StringVarP(&folder, "folder", "f", "", "Folder to create the job in")
	cmd.Flags().BoolVar(&ifNotExists, "if-not-exists", false, "Don't error if the job already exists")

	return cmd
}
