package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newSystemQuietDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "quiet-down",
		Short:       "Put Jenkins into quiet-down mode",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Put Jenkins into quiet-down mode. No new builds will be started.

In quiet-down mode, Jenkins will not start any new builds. Builds
already running will continue. This is useful before a planned
restart or maintenance window.

Examples:
  # Enter quiet-down mode
  jenkins system quiet-down`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := jenkinsClient.QuietDown(); err != nil {
				return fmt.Errorf("quieting down: %w", err)
			}

			fmt.Fprintln(os.Stdout, "Jenkins is now in quiet-down mode. No new builds will start.")
			return nil
		},
	}
}
