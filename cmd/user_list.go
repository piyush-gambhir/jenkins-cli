package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newUserListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List users",
		Long: `List all known Jenkins users.

Displays each user's ID, full name, and last activity timestamp.

Examples:
  # List all users
  jenkins user list

  # Output as JSON
  jenkins user list -o json`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			users, err := jenkinsClient.ListUsers()
			if err != nil {
				return fmt.Errorf("listing users: %w", err)
			}

			if len(users) == 0 {
				fmt.Fprintln(os.Stdout, "No users found.")
				return nil
			}

			tableDef := &output.TableDef{
				Headers: []string{"ID", "NAME", "LAST ACTIVITY"},
				RowFunc: func(item interface{}) []string {
					u := item.(client.UserListItem)
					lastChange := "N/A"
					if u.LastChange > 0 {
						lastChange = client.FormatTimestamp(u.LastChange)
					}
					return []string{
						u.User.ID,
						u.User.FullName,
						lastChange,
					}
				},
			}

			return output.Print(os.Stdout, outFormat, users, tableDef)
		},
	}
}
