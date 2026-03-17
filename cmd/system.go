package cmd

import (
	"github.com/spf13/cobra"
)

func newSystemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "System administration commands",
		Long:  "Jenkins system administration operations including restart, quiet-down, and script execution.",
	}

	cmd.AddCommand(newSystemInfoCmd())
	cmd.AddCommand(newSystemRestartCmd())
	cmd.AddCommand(newSystemQuietDownCmd())
	cmd.AddCommand(newSystemCancelQuietDownCmd())
	cmd.AddCommand(newSystemRunScriptCmd())

	return cmd
}
