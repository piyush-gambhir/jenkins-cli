package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newPluginUninstallCmd() *cobra.Command {
	var confirm bool
	var ifExists bool

	cmd := &cobra.Command{
		Use:         "uninstall <plugin-name>",
		Short:       "Uninstall a plugin",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Uninstall a Jenkins plugin. A restart will be required.

Marks the plugin for removal. Jenkins must be restarted for the change
to take effect. Requires --confirm.

Examples:
  # Uninstall a plugin
  jenkins plugin uninstall git --confirm

  # Idempotent uninstall (no error if plugin isn't installed)
  jenkins plugin uninstall git --confirm --if-exists`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !confirm {
				if noInputFlag {
					return fmt.Errorf("interactive input required but --no-input is set. Use --confirm for destructive operations.")
				}
				return fmt.Errorf("use --confirm to confirm uninstallation of plugin %q", name)
			}

			if err := jenkinsClient.UninstallPlugin(name); err != nil {
				var apiErr *client.APIError
				if ifExists && errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
					if !quietFlag {
						fmt.Fprintf(os.Stdout, "Plugin %q is not installed, skipping.\n", name)
					}
					return nil
				}
				return fmt.Errorf("uninstalling plugin: %w", err)
			}

			if !quietFlag {
				fmt.Fprintf(os.Stdout, "Plugin %q marked for uninstallation. Restart Jenkins to complete.\n", name)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm uninstallation")
	cmd.Flags().BoolVar(&ifExists, "if-exists", false, "Don't error if the plugin isn't installed")

	return cmd
}
