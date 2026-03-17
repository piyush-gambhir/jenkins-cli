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
		Use:   "create <job-name>",
		Short: "Create a new job",
		Long:  "Create a new Jenkins job from an XML configuration file.",
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
