package cmd

import (
	"github.com/spf13/cobra"
)

func newViewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "view",
		Aliases: []string{"views"},
		Short:   "Manage Jenkins views",
		Long: `List, create, and manage Jenkins views.

Views are named collections of jobs displayed on the Jenkins dashboard.

Subcommands:
  list         List all views
  get          Get detailed info about a view including its jobs
  create       Create a new view
  delete       Delete a view
  add-job      Add a job to a view
  remove-job   Remove a job from a view`,
	}

	cmd.AddCommand(newViewListCmd())
	cmd.AddCommand(newViewGetCmd())
	cmd.AddCommand(newViewCreateCmd())
	cmd.AddCommand(newViewDeleteCmd())
	cmd.AddCommand(newViewAddJobCmd())
	cmd.AddCommand(newViewRemoveJobCmd())

	return cmd
}
