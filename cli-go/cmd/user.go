package cmd

import (
	"github.com/spf13/cobra"
)

func newUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "user",
		Aliases: []string{"users"},
		Short:   "Manage Jenkins users",
		Long: `List and inspect Jenkins users.

Subcommands:
  list   List all known users
  get    Get details about a specific user`,
	}

	cmd.AddCommand(newUserListCmd())
	cmd.AddCommand(newUserGetCmd())

	return cmd
}
