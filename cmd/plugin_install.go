package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newPluginInstallCmd() *cobra.Command {
	var version string

	cmd := &cobra.Command{
		Use:   "install <plugin-name>",
		Short: "Install a plugin",
		Long:  "Install a Jenkins plugin by name. Optionally specify a version.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := jenkinsClient.InstallPlugin(name, version); err != nil {
				return fmt.Errorf("installing plugin: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Plugin %q installation initiated. A restart may be required.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Plugin version to install")

	return cmd
}
