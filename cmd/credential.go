package cmd

import (
	"github.com/spf13/cobra"
)

func newCredentialCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "credential",
		Aliases: []string{"credentials", "cred", "creds"},
		Short:   "Manage Jenkins credentials",
		Long:    "List, create, and manage Jenkins credentials.",
	}

	cmd.AddCommand(newCredentialListCmd())
	cmd.AddCommand(newCredentialGetCmd())
	cmd.AddCommand(newCredentialCreateCmd())
	cmd.AddCommand(newCredentialUpdateCmd())
	cmd.AddCommand(newCredentialDeleteCmd())

	return cmd
}
