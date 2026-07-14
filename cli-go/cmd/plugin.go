package cmd

import (
	"github.com/spf13/cobra"
)

func newPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugin",
		Aliases: []string{"plugins"},
		Short:   "Manage Jenkins plugins",
		Long: `List, install, and manage Jenkins plugins.

Subcommands:
  list            List installed plugins (with optional --active / --enabled filters)
  get             Get detailed info about a specific plugin
  install         Install a plugin by name
  uninstall       Uninstall a plugin
  check-updates   Check for available plugin updates`,
	}

	cmd.AddCommand(newPluginListCmd())
	cmd.AddCommand(newPluginGetCmd())
	cmd.AddCommand(newPluginInstallCmd())
	cmd.AddCommand(newPluginUninstallCmd())
	cmd.AddCommand(newPluginCheckUpdatesCmd())

	return cmd
}
