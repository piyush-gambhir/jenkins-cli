package cmd

import (
	"github.com/spf13/cobra"
)

func newQueueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue",
		Short: "Manage the build queue",
		Long:  "List and manage items in the Jenkins build queue.",
	}

	cmd.AddCommand(newQueueListCmd())
	cmd.AddCommand(newQueueCancelCmd())

	return cmd
}
