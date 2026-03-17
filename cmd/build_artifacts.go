package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newBuildArtifactsCmd() *cobra.Command {
	var download bool
	var outputDir string

	cmd := &cobra.Command{
		Use:   "artifacts <job-path> <build-number>",
		Short: "List or download build artifacts",
		Long: `List artifacts for a build. Use --download to download them.

By default, displays a table of artifact file names and paths. Use
--download to save all artifacts to the current directory (or specify
--output-dir for a custom location).

Examples:
  # List artifacts of build #42
  jenkins build artifacts my-pipeline 42

  # Download all artifacts to the current directory
  jenkins build artifacts my-pipeline 42 --download

  # Download artifacts to a specific directory
  jenkins build artifacts my-pipeline 42 --download --output-dir ./artifacts

  # List artifacts as JSON
  jenkins build artifacts my-pipeline 42 -o json`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			number, err := client.ParseBuildNumber(args[1])
			if err != nil {
				return err
			}

			artifacts, err := jenkinsClient.GetBuildArtifacts(jobPath, number)
			if err != nil {
				return fmt.Errorf("getting artifacts: %w", err)
			}

			if len(artifacts) == 0 {
				fmt.Fprintln(os.Stdout, "No artifacts found.")
				return nil
			}

			if download {
				if outputDir == "" {
					outputDir = "."
				}
				for _, a := range artifacts {
					data, err := jenkinsClient.DownloadArtifact(jobPath, number, a.RelativePath)
					if err != nil {
						return fmt.Errorf("downloading %s: %w", a.FileName, err)
					}
					outPath := filepath.Join(outputDir, a.FileName)
					if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
						return fmt.Errorf("creating directory: %w", err)
					}
					if err := os.WriteFile(outPath, data, 0o644); err != nil {
						return fmt.Errorf("writing %s: %w", outPath, err)
					}
					fmt.Fprintf(os.Stdout, "Downloaded: %s\n", outPath)
				}
				return nil
			}

			tableDef := &output.TableDef{
				Headers: []string{"FILE NAME", "RELATIVE PATH"},
				RowFunc: func(item interface{}) []string {
					a := item.(client.Artifact)
					return []string{a.FileName, a.RelativePath}
				},
			}

			return output.Print(os.Stdout, outFormat, artifacts, tableDef)
		},
	}

	cmd.Flags().BoolVarP(&download, "download", "d", false, "Download artifacts")
	cmd.Flags().StringVar(&outputDir, "output-dir", "", "Directory to download artifacts to")

	return cmd
}
