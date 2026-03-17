package cmd

import (
	"github.com/spf13/cobra"
)

func newViewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "view",
		Aliases: []string{"views"},
		Short:   "Manage Jenkins views",
		Long:    "List, create, and manage Jenkins views.",
	}

	cmd.AddCommand(newViewListCmd())
	cmd.AddCommand(newViewGetCmd())
	cmd.AddCommand(newViewCreateCmd())
	cmd.AddCommand(newViewDeleteCmd())
	cmd.AddCommand(newViewAddJobCmd())
	cmd.AddCommand(newViewRemoveJobCmd())

	return cmd
}
