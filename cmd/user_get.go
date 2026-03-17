package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newUserGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <user-id>",
		Short: "Get user details",
		Long: `Display details about a specific Jenkins user.

Shows the user's ID, full name, URL, and description.

Examples:
  # Get user details
  jenkins user get admin

  # Output as JSON
  jenkins user get admin -o json`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			user, err := jenkinsClient.GetUser(id)
			if err != nil {
				return fmt.Errorf("getting user: %w", err)
			}

			if outFormat == output.FormatTable {
				fmt.Fprintf(os.Stdout, "User: %s\n", user.ID)
				fmt.Fprintf(os.Stdout, "  Name:  %s\n", user.FullName)
				fmt.Fprintf(os.Stdout, "  URL:   %s\n", user.AbsoluteURL)
				if user.Description != "" {
					fmt.Fprintf(os.Stdout, "  About: %s\n", user.Description)
				}
				return nil
			}

			return output.Print(os.Stdout, outFormat, user, nil)
		},
	}
}
