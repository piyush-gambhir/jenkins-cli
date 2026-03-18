package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newPluginInstallCmd() *cobra.Command {
	var version string
	var ifNotExists bool

	cmd := &cobra.Command{
		Use:         "install <plugin-name>",
		Short:       "Install a plugin",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Install a Jenkins plugin by name. Optionally specify a version.

Initiates plugin installation on the Jenkins server. A restart may be
required for the plugin to become active. Use --version to install a
specific version.

Examples:
  # Install the latest version of the git plugin
  jenkins plugin install git

  # Install a specific version
  jenkins plugin install git --version 5.2.0

  # Idempotent install (no error if plugin is already installed)
  jenkins plugin install git --if-not-exists`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := jenkinsClient.InstallPlugin(name, version); err != nil {
				var apiErr *client.APIError
				if ifNotExists && errors.As(err, &apiErr) && (apiErr.StatusCode == 400 || apiErr.StatusCode == 409) {
					if !quietFlag {
						fmt.Fprintf(os.Stdout, "Plugin %q is already installed, skipping.\n", name)
					}
					return nil
				}
				return fmt.Errorf("installing plugin: %w", err)
			}

			if !quietFlag {
				fmt.Fprintf(os.Stdout, "Plugin %q installation initiated. A restart may be required.\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Plugin version to install")
	cmd.Flags().BoolVar(&ifNotExists, "if-not-exists", false, "Don't error if the plugin is already installed")

	return cmd
}
