package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newWhoAmICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current user information",
		Long:  "Display information about the currently authenticated Jenkins user.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := jenkinsClient.WhoAmI()
			if err != nil {
				return fmt.Errorf("getting user info: %w", err)
			}

			if outFormat == output.FormatTable {
				fmt.Fprintf(os.Stdout, "User Information\n")
				fmt.Fprintf(os.Stdout, "  ID:       %s\n", user.ID)
				fmt.Fprintf(os.Stdout, "  Name:     %s\n", user.FullName)
				fmt.Fprintf(os.Stdout, "  URL:      %s\n", user.AbsoluteURL)
				if user.Description != "" {
					fmt.Fprintf(os.Stdout, "  About:    %s\n", user.Description)
				}
				return nil
			}

			return output.Print(os.Stdout, outFormat, user, nil)
		},
	}
}
