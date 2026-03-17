package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/config"
	"github.com/piyush-gambhir/jenkins-cli/internal/update"
	"github.com/piyush-gambhir/jenkins-cli/internal/version"
)

func newUpdateCmd() *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update jenkins to the latest version",
		Long:  "Check for and install the latest version of the Jenkins CLI from GitHub Releases.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := "piyush-gambhir/jenkins-cli"
			currentVersion := version.Version

			if currentVersion == "dev" {
				return fmt.Errorf("cannot update a dev build. Install a release version from https://github.com/%s/releases", repo)
			}

			fmt.Fprint(os.Stderr, "Checking for updates... ")
			info, err := update.CheckForUpdate(currentVersion, repo, config.ConfigDir(), true)
			if err != nil {
				fmt.Fprintln(os.Stderr, "")
				return fmt.Errorf("checking for updates: %w", err)
			}
			fmt.Fprintln(os.Stderr, "done.")

			if !info.Available {
				fmt.Printf("Already up to date (v%s)\n", currentVersion)
				return nil
			}

			fmt.Printf("Update available: v%s → v%s\n", info.CurrentVersion, info.LatestVersion)
			if info.ReleaseURL != "" {
				fmt.Printf("Release: %s\n", info.ReleaseURL)
			}

			if checkOnly {
				return nil
			}

			// Ask for confirmation
			fmt.Printf("\nDo you want to update? [y/N] ")
			var response string
			fmt.Scanln(&response)
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Update cancelled.")
				return nil
			}

			// Determine download URL
			osName := runtime.GOOS
			archName := runtime.GOARCH
			archive := fmt.Sprintf("jenkins-cli_%s_%s.tar.gz", osName, archName)
			downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s",
				repo, info.LatestVersion, archive)

			// Find current binary path
			execPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("finding current binary: %w", err)
			}
			execPath, err = filepath.EvalSymlinks(execPath)
			if err != nil {
				return fmt.Errorf("resolving binary path: %w", err)
			}

			fmt.Printf("Downloading v%s...\n", info.LatestVersion)

			// Download to temp directory
			tmpDir, err := os.MkdirTemp("", "jenkins-cli-update-*")
			if err != nil {
				return fmt.Errorf("creating temp directory: %w", err)
			}
			defer os.RemoveAll(tmpDir)

			archivePath := filepath.Join(tmpDir, archive)
			if err := downloadFile(downloadURL, archivePath); err != nil {
				return fmt.Errorf("downloading update: %w", err)
			}

			// Extract binary from tar.gz
			fmt.Println("Extracting...")
			binaryPath := filepath.Join(tmpDir, "jenkins")
			if err := extractBinary(archivePath, binaryPath); err != nil {
				return fmt.Errorf("extracting update: %w", err)
			}

			// Check if we can write to the destination
			fmt.Printf("Installing to %s...\n", execPath)

			// Get permissions of the existing binary
			existingStat, err := os.Stat(execPath)
			if err != nil {
				return fmt.Errorf("checking binary permissions: %w", err)
			}

			// Try atomic replace: rename new binary over old one.
			// First, copy to a temp file in the same directory as the target
			// (rename only works within the same filesystem).
			targetDir := filepath.Dir(execPath)
			tmpBin, err := os.CreateTemp(targetDir, ".jenkins-update-*")
			if err != nil {
				// If we can't write to the target directory, suggest sudo
				return fmt.Errorf("cannot write to %s: %w\nTry: sudo jenkins update", targetDir, err)
			}
			tmpBinPath := tmpBin.Name()

			// Copy new binary to temp location in target dir
			src, err := os.Open(binaryPath)
			if err != nil {
				os.Remove(tmpBinPath)
				return fmt.Errorf("opening new binary: %w", err)
			}

			if _, err := io.Copy(tmpBin, src); err != nil {
				src.Close()
				tmpBin.Close()
				os.Remove(tmpBinPath)
				return fmt.Errorf("copying new binary: %w", err)
			}
			src.Close()
			tmpBin.Close()

			// Set permissions to match the original binary
			if err := os.Chmod(tmpBinPath, existingStat.Mode()); err != nil {
				os.Remove(tmpBinPath)
				return fmt.Errorf("setting permissions: %w", err)
			}

			// Atomic rename
			if err := os.Rename(tmpBinPath, execPath); err != nil {
				os.Remove(tmpBinPath)
				return fmt.Errorf("replacing binary: %w\nTry: sudo jenkins update", err)
			}

			fmt.Printf("Successfully updated jenkins to v%s\n", info.LatestVersion)
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check if an update is available, don't install")

	return cmd
}

// downloadFile downloads a URL to a local file.
func downloadFile(url, dest string) error {
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d for %s", resp.StatusCode, url)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractBinary extracts the "jenkins" binary from a tar.gz archive.
func extractBinary(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("opening gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tar: %w", err)
		}

		// Look for the jenkins binary (could be at root or in a subdirectory)
		name := filepath.Base(header.Name)
		if name == "jenkins" && header.Typeflag == tar.TypeReg {
			out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
			return nil
		}
	}

	return fmt.Errorf("binary 'jenkins' not found in archive")
}
