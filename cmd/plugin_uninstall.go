package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newPluginUninstallCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "uninstall <plugin-name>",
		Short: "Uninstall a plugin",
		Long:  "Uninstall a Jenkins plugin. A restart will be required.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !confirm {
				return fmt.Errorf("use --confirm to confirm uninstallation of plugin %q", name)
			}

			if err := jenkinsClient.UninstallPlugin(name); err != nil {
				return fmt.Errorf("uninstalling plugin: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Plugin %q marked for uninstallation. Restart Jenkins to complete.\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm uninstallation")

	return cmd
}
