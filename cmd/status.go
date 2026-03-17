package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Jenkins server status",
		Long: `Display information about the connected Jenkins server including version, mode, security, and executor count.

Shows the server URL, version, mode, security settings, executor count,
quiet-down status, and description.

Examples:
  # Show server status
  jenkins status

  # Output as JSON
  jenkins status -o json

  # Use a specific profile
  jenkins status --profile staging`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := jenkinsClient.GetServerInfo()
			if err != nil {
				return fmt.Errorf("getting server status: %w", err)
			}

			version, err := jenkinsClient.GetServerVersion()
			if err != nil {
				version = "unknown"
			}

			if outFormat == output.FormatTable {
				fmt.Fprintf(os.Stdout, "Jenkins Server Status\n")
				fmt.Fprintf(os.Stdout, "  URL:           %s\n", jenkinsClient.BaseURL())
				fmt.Fprintf(os.Stdout, "  Version:       %s\n", version)
				fmt.Fprintf(os.Stdout, "  Mode:          %s\n", info.Mode)
				fmt.Fprintf(os.Stdout, "  Security:      %v\n", info.UseSecurity)
				fmt.Fprintf(os.Stdout, "  CSRF:          %v\n", info.UseCrumbs)
				fmt.Fprintf(os.Stdout, "  Executors:     %d\n", info.NumExecutors)
				fmt.Fprintf(os.Stdout, "  Quieting Down: %v\n", info.QuietingDown)
				if info.Description != "" {
					fmt.Fprintf(os.Stdout, "  Description:   %s\n", info.Description)
				}
				return nil
			}

			result := map[string]interface{}{
				"url":          jenkinsClient.BaseURL(),
				"version":      version,
				"mode":         info.Mode,
				"security":     info.UseSecurity,
				"csrf":         info.UseCrumbs,
				"executors":    info.NumExecutors,
				"quietingDown": info.QuietingDown,
				"description":  info.Description,
			}

			return output.Print(os.Stdout, outFormat, result, nil)
		},
	}
}
