package cmd

import (
	"github.com/spf13/cobra"
)

func newSystemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "System administration commands",
		Long: `Jenkins system administration operations including restart, quiet-down, and script execution.

Subcommands:
  info                Show Jenkins system info (version, mode, executors, etc.)
  restart             Restart Jenkins (with optional --safe for graceful restart)
  quiet-down          Enter quiet-down mode (no new builds start)
  cancel-quiet-down   Exit quiet-down mode
  run-script          Execute a Groovy script on the Jenkins controller`,
	}

	cmd.AddCommand(newSystemInfoCmd())
	cmd.AddCommand(newSystemRestartCmd())
	cmd.AddCommand(newSystemQuietDownCmd())
	cmd.AddCommand(newSystemCancelQuietDownCmd())
	cmd.AddCommand(newSystemRunScriptCmd())

	return cmd
}
