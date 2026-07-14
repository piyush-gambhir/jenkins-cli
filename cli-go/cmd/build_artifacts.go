package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/output"
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
		Args: cobra.ExactArgs(2),
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

			if len(artifacts) == 0 && outFormat == output.FormatTable {
				if !quietFlag {
					fmt.Fprintln(os.Stdout, "No artifacts found.")
				}
				return nil
			}

			if download {
				if outputDir == "" {
					outputDir = "."
				}
				for _, a := range artifacts {
					outPath, err := artifactOutputPath(outputDir, a.RelativePath, a.FileName)
					if err != nil {
						return fmt.Errorf("invalid artifact path %q: %w", a.RelativePath, err)
					}
					if err := prepareArtifactPath(outputDir, outPath); err != nil {
						return fmt.Errorf("invalid artifact path %q: %w", a.RelativePath, err)
					}
					tmp, err := os.CreateTemp(filepath.Dir(outPath), ".artifact-*.tmp")
					if err != nil {
						return fmt.Errorf("creating artifact file: %w", err)
					}
					tmpName := tmp.Name()
					if err := jenkinsClient.DownloadArtifactTo(cmd.Context(), jobPath, number, a.RelativePath, tmp); err != nil {
						tmp.Close()
						os.Remove(tmpName)
						return fmt.Errorf("downloading %s: %w", a.FileName, err)
					}
					if err := tmp.Chmod(0o644); err != nil {
						tmp.Close()
						os.Remove(tmpName)
						return fmt.Errorf("setting artifact permissions: %w", err)
					}
					if err := tmp.Sync(); err != nil {
						tmp.Close()
						os.Remove(tmpName)
						return fmt.Errorf("syncing artifact: %w", err)
					}
					if err := tmp.Close(); err != nil {
						os.Remove(tmpName)
						return fmt.Errorf("closing artifact: %w", err)
					}
					if err := os.Rename(tmpName, outPath); err != nil {
						os.Remove(tmpName)
						return fmt.Errorf("replacing %s: %w", outPath, err)
					}
					if !quietFlag {
						fmt.Fprintf(os.Stdout, "Downloaded: %s\n", outPath)
					}
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

func prepareArtifactPath(outputDir, outPath string) error {
	root, err := filepath.Abs(outputDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	parent := filepath.Dir(outPath)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	realParent, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(realRoot, realParent)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("parent directory escapes output directory through a symlink")
	}
	if info, err := os.Lstat(outPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("destination is a symbolic link")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func artifactOutputPath(outputDir, relativePath, fileName string) (string, error) {
	rel := relativePath
	if rel == "" {
		rel = fileName
	}
	rel = filepath.Clean(filepath.FromSlash(rel))
	if rel == "." || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes output directory")
	}

	root, err := filepath.Abs(outputDir)
	if err != nil {
		return "", err
	}
	out := filepath.Join(root, rel)
	within, err := filepath.Rel(root, out)
	if err != nil || within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes output directory")
	}
	return out, nil
}
