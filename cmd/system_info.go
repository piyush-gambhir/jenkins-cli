package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newSystemInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show system information",
		Long:  "Display detailed Jenkins system information.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := jenkinsClient.GetServerInfo()
			if err != nil {
				return fmt.Errorf("getting system info: %w", err)
			}

			version, err := jenkinsClient.GetServerVersion()
			if err != nil {
				version = "unknown"
			}

			if outFormat == output.FormatTable {
				fmt.Fprintf(os.Stdout, "Jenkins System Information\n")
				fmt.Fprintf(os.Stdout, "  URL:             %s\n", jenkinsClient.BaseURL())
				fmt.Fprintf(os.Stdout, "  Version:         %s\n", version)
				fmt.Fprintf(os.Stdout, "  Mode:            %s\n", info.Mode)
				fmt.Fprintf(os.Stdout, "  Description:     %s\n", info.NodeDescription)
				fmt.Fprintf(os.Stdout, "  Executors:       %d\n", info.NumExecutors)
				fmt.Fprintf(os.Stdout, "  Security:        %v\n", info.UseSecurity)
				fmt.Fprintf(os.Stdout, "  CSRF:            %v\n", info.UseCrumbs)
				fmt.Fprintf(os.Stdout, "  Quieting Down:   %v\n", info.QuietingDown)
				if info.PrimaryView != nil {
					fmt.Fprintf(os.Stdout, "  Primary View:    %s\n", info.PrimaryView.Name)
				}
				fmt.Fprintf(os.Stdout, "  Total Views:     %d\n", len(info.Views))
				return nil
			}

			result := map[string]interface{}{
				"url":          jenkinsClient.BaseURL(),
				"version":      version,
				"mode":         info.Mode,
				"description":  info.NodeDescription,
				"executors":    info.NumExecutors,
				"security":     info.UseSecurity,
				"csrf":         info.UseCrumbs,
				"quietingDown": info.QuietingDown,
				"views":        len(info.Views),
			}

			return output.Print(os.Stdout, outFormat, result, nil)
		},
	}
}
