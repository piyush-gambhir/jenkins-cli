package cmd

import (
	"github.com/spf13/cobra"
)

func newPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugin",
		Aliases: []string{"plugins"},
		Short:   "Manage Jenkins plugins",
		Long:    "List, install, and manage Jenkins plugins.",
	}

	cmd.AddCommand(newPluginListCmd())
	cmd.AddCommand(newPluginGetCmd())
	cmd.AddCommand(newPluginInstallCmd())
	cmd.AddCommand(newPluginUninstallCmd())
	cmd.AddCommand(newPluginCheckUpdatesCmd())

	return cmd
}
