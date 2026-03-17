package cmd

import (
	"github.com/spf13/cobra"
)

func newQueueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue",
		Short: "Manage the build queue",
		Long: `List and manage items in the Jenkins build queue.

Subcommands:
  list     List all items currently waiting in the build queue
  cancel   Cancel a pending build by its queue ID`,
	}

	cmd.AddCommand(newQueueListCmd())
	cmd.AddCommand(newQueueCancelCmd())

	return cmd
}
