package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newNodeEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "enable <node-name>",
		Short:       "Bring a node online",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Bring an offline Jenkins node back online.

Toggles a temporarily offline node back to online status so it can
accept build tasks again.

Examples:
  # Bring a node online
  jenkins node enable my-agent`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := jenkinsClient.ToggleOffline(name, false, ""); err != nil {
				return fmt.Errorf("enabling node: %w", err)
			}

			if !quietFlag {
				fmt.Fprintf(os.Stdout, "Node %q brought online.\n", name)
			}
			return nil
		},
	}
}
