package cmd

import (
	"github.com/spf13/cobra"
)

func newCredentialCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "credential",
		Aliases: []string{"credentials", "cred", "creds"},
		Short:   "Manage Jenkins credentials",
		Long: `List, create, and manage Jenkins credentials.

Jenkins stores credentials in stores and domains. The default store is
"system" and the default domain is "_" (global). Most commands accept
--store and --domain flags.

Subcommands:
  list     List credentials (with optional --type filter)
  get      Get details about a specific credential
  create   Create a credential from XML config
  update   Update a credential from XML config
  delete   Delete a credential`,
	}

	cmd.AddCommand(newCredentialListCmd())
	cmd.AddCommand(newCredentialGetCmd())
	cmd.AddCommand(newCredentialCreateCmd())
	cmd.AddCommand(newCredentialUpdateCmd())
	cmd.AddCommand(newCredentialDeleteCmd())

	return cmd
}
