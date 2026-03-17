package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newBuildTestReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test-report <job-path> <build-number>",
		Short: "Get build test report",
		Long: `Display the test report for a build including pass/fail/skip counts.

Shows total, passed, failed, and skipped test counts along with duration.
If there are failures, individual failed test cases are listed with their
error messages.

Examples:
  # View test report for build #42
  jenkins build test-report my-pipeline 42

  # View test report as JSON (for parsing)
  jenkins build test-report my-pipeline 42 -o json

  # View test report for a job in a folder
  jenkins build test-report my-folder/my-pipeline 10`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			number, err := client.ParseBuildNumber(args[1])
			if err != nil {
				return err
			}

			report, err := jenkinsClient.GetBuildTestReport(jobPath, number)
			if err != nil {
				return fmt.Errorf("getting test report: %w", err)
			}

			if outFormat == output.FormatTable {
				fmt.Fprintf(os.Stdout, "Test Report for Build #%d\n", number)
				fmt.Fprintf(os.Stdout, "  Total:    %d\n", report.TotalCount)
				fmt.Fprintf(os.Stdout, "  Passed:   %d\n", report.PassCount)
				fmt.Fprintf(os.Stdout, "  Failed:   %d\n", report.FailCount)
				fmt.Fprintf(os.Stdout, "  Skipped:  %d\n", report.SkipCount)
				fmt.Fprintf(os.Stdout, "  Duration: %.2fs\n", report.Duration)

				if report.FailCount > 0 {
					fmt.Fprintf(os.Stdout, "\nFailed Tests:\n")
					for _, suite := range report.Suites {
						for _, tc := range suite.Cases {
							if tc.Status == "FAILED" || tc.Status == "REGRESSION" {
								fmt.Fprintf(os.Stdout, "  - %s.%s\n", tc.ClassName, tc.Name)
								if tc.ErrorMsg != "" {
									fmt.Fprintf(os.Stdout, "    Error: %s\n", tc.ErrorMsg)
								}
							}
						}
					}
				}
				return nil
			}

			return output.Print(os.Stdout, outFormat, report, nil)
		},
	}
}
