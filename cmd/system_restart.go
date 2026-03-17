package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newSystemRestartCmd() *cobra.Command {
	var safe bool
	var confirm bool

	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart Jenkins",
		Long:  "Restart the Jenkins server. Use --safe to wait for running builds to complete.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return fmt.Errorf("use --confirm to confirm restarting Jenkins")
			}

			if safe {
				if err := jenkinsClient.SafeRestart(); err != nil {
					return fmt.Errorf("safe restarting: %w", err)
				}
				fmt.Fprintln(os.Stdout, "Jenkins safe restart initiated. Waiting for running builds to complete.")
			} else {
				if err := jenkinsClient.Restart(); err != nil {
					return fmt.Errorf("restarting: %w", err)
				}
				fmt.Fprintln(os.Stdout, "Jenkins restart initiated.")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&safe, "safe", false, "Safe restart (wait for builds)")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm restart")

	return cmd
}
