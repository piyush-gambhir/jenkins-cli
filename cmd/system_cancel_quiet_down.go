package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newSystemCancelQuietDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel-quiet-down",
		Short: "Cancel quiet-down mode",
		Long: `Cancel Jenkins quiet-down mode, resuming normal operations.

Exits quiet-down mode so Jenkins will resume starting new builds.

Examples:
  # Cancel quiet-down mode
  jenkins system cancel-quiet-down`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := jenkinsClient.CancelQuietDown(); err != nil {
				return fmt.Errorf("cancelling quiet down: %w", err)
			}

			fmt.Fprintln(os.Stdout, "Quiet-down mode cancelled. Jenkins is accepting new builds.")
			return nil
		},
	}
}
