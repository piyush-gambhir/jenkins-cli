package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newQueueCancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "cancel <queue-id>",
		Short:       "Cancel a queued item",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Cancel a pending build in the queue by its ID.

Use "jenkins queue list" to find the queue item ID, then pass it to
this command to cancel the pending build.

Examples:
  # Cancel queue item with ID 123
  jenkins queue cancel 123`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseNumber(args[0])
			if err != nil {
				return fmt.Errorf("invalid queue ID: %w", err)
			}

			if err := jenkinsClient.CancelQueueItem(id); err != nil {
				return fmt.Errorf("cancelling queue item: %w", err)
			}

			if !quietFlag {
				fmt.Fprintf(os.Stdout, "Queue item %d cancelled.\n", id)
			}
			return nil
		},
	}
}
