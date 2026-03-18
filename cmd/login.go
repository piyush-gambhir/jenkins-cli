package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/config"
)

func newLoginCmd() *cobra.Command {
	var profileName string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a Jenkins server",
		Long: `Interactively configure a Jenkins server connection profile.

Prompts for the Jenkins URL, username, and API token, then tests the
connection. The profile is saved to the config file for future use.

Jenkins requires an API token (not your password). Generate one at:
  <jenkins-url>/user/<username>/configure  (API Token section)

Examples:
  # Interactive login (prompts for all values)
  jenkins login

  # Login and save with a specific profile name
  jenkins login --name staging`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if noInputFlag {
				return fmt.Errorf("interactive input required but --no-input is set. Use environment variables JENKINS_URL, JENKINS_USER, and JENKINS_TOKEN instead.")
			}

			reader := bufio.NewReader(os.Stdin)

			// Prompt URL
			fmt.Print("Jenkins URL: ")
			urlStr, _ := reader.ReadString('\n')
			urlStr = strings.TrimSpace(urlStr)
			if urlStr == "" {
				return fmt.Errorf("URL is required")
			}

			// Prompt username
			fmt.Print("Username: ")
			username, _ := reader.ReadString('\n')
			username = strings.TrimSpace(username)
			if username == "" {
				return fmt.Errorf("username is required")
			}

			// Prompt token
			fmt.Print("API Token: ")
			token, _ := reader.ReadString('\n')
			token = strings.TrimSpace(token)
			if token == "" {
				return fmt.Errorf("API token is required")
			}

			// Prompt profile name
			if profileName == "" {
				fmt.Print("Profile name [default]: ")
				profileName, _ = reader.ReadString('\n')
				profileName = strings.TrimSpace(profileName)
				if profileName == "" {
					profileName = "default"
				}
			}

			// Test connection
			profile := config.Profile{
				URL:      strings.TrimRight(urlStr, "/"),
				Username: username,
				Token:    token,
				Insecure: insecureFlag,
			}

			if !quietFlag {
				fmt.Print("Testing connection... ")
			}
			c := client.NewClient(profile)
			user, err := c.WhoAmI()
			if err != nil {
				if !quietFlag {
					fmt.Println("FAILED")
				}
				return fmt.Errorf("connection test failed: %w", err)
			}
			if !quietFlag {
				fmt.Printf("OK (authenticated as %s)\n", user.FullName)
			}

			// Load config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			// Save profile
			config.SetProfile(cfg, profileName, profile)
			if cfg.CurrentProfile == "" {
				cfg.CurrentProfile = profileName
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			if !quietFlag {
				fmt.Printf("Profile %q saved to %s\n", profileName, config.ConfigPath())
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&profileName, "name", "", "Profile name (default: prompted)")

	return cmd
}
