package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newQueueCancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <queue-id>",
		Short: "Cancel a queued item",
		Long:  "Cancel a pending build in the queue by its ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseNumber(args[0])
			if err != nil {
				return fmt.Errorf("invalid queue ID: %w", err)
			}

			if err := jenkinsClient.CancelQueueItem(id); err != nil {
				return fmt.Errorf("cancelling queue item: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Queue item %d cancelled.\n", id)
			return nil
		},
	}
}
