package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	jpath "github.com/piyush-gambhir/jenkins-cli/internal/path"
)

func newBuildOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open <job-path> <build-number>",
		Short: "Open build in browser",
		Long:  "Open the build page in your default web browser.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			number, err := client.ParseBuildNumber(args[1])
			if err != nil {
				return err
			}

			jp := jpath.ToJenkinsPath(jobPath)
			buildURL := fmt.Sprintf("%s%s/%d", jenkinsClient.BaseURL(), jp, number)

			fmt.Fprintf(os.Stdout, "Opening %s\n", buildURL)

			return openBrowser(buildURL)
		},
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
	return cmd.Start()
}
