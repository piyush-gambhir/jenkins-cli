package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newJobGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <job-path>",
		Short: "Get job details",
		Long:  "Display detailed information about a specific job.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]

			job, err := jenkinsClient.GetJob(jobPath)
			if err != nil {
				return fmt.Errorf("getting job: %w", err)
			}

			if outFormat == output.FormatTable {
				fmt.Fprintf(os.Stdout, "Job: %s\n", job.Name)
				fmt.Fprintf(os.Stdout, "  Full Name:   %s\n", job.FullName)
				fmt.Fprintf(os.Stdout, "  URL:         %s\n", job.URL)
				fmt.Fprintf(os.Stdout, "  Buildable:   %v\n", job.Buildable)
				fmt.Fprintf(os.Stdout, "  In Queue:    %v\n", job.InQueue)
				fmt.Fprintf(os.Stdout, "  Color:       %s\n", job.Color)
				if job.Description != "" {
					fmt.Fprintf(os.Stdout, "  Description: %s\n", job.Description)
				}
				if job.LastBuild != nil {
					fmt.Fprintf(os.Stdout, "  Last Build:  #%d (%s)\n", job.LastBuild.Number, job.LastBuild.Result)
				}
				for _, hr := range job.HealthReport {
					fmt.Fprintf(os.Stdout, "  Health:      %s (score: %d%%)\n", hr.Description, hr.Score)
				}

				// Show parameters if any
				for _, prop := range job.Property {
					if len(prop.ParameterDefinitions) > 0 {
						fmt.Fprintf(os.Stdout, "  Parameters:\n")
						for _, p := range prop.ParameterDefinitions {
							fmt.Fprintf(os.Stdout, "    - %s (%s)\n", p.Name, p.Type)
							if p.Description != "" {
								fmt.Fprintf(os.Stdout, "      %s\n", p.Description)
							}
						}
					}
				}

				// Show child jobs if folder
				if len(job.Jobs) > 0 {
					fmt.Fprintf(os.Stdout, "  Jobs:\n")
					for _, j := range job.Jobs {
						fmt.Fprintf(os.Stdout, "    - %s (%s)\n", j.Name, client.ColorToStatus(j.Color))
					}
				}
				return nil
			}

			return output.Print(os.Stdout, outFormat, job, nil)
		},
	}
}
