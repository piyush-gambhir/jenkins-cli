package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newNodeDisableCmd() *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "disable <node-name>",
		Short: "Take a node offline",
		Long:  "Take a Jenkins node offline with an optional message.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := jenkinsClient.ToggleOffline(name, true, message); err != nil {
				return fmt.Errorf("disabling node: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Node %q taken offline.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Offline reason message")

	return cmd
}
