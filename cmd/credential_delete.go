package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCredentialDeleteCmd() *cobra.Command {
	var store string
	var domain string
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <credential-id>",
		Short: "Delete a credential",
		Long: `Delete a Jenkins credential.

Permanently removes a credential from the specified store and domain.
Requires --confirm.

WARNING: Any jobs referencing this credential will fail on their next run.

Examples:
  # Delete a credential
  jenkins credential delete my-cred-id --confirm

  # Delete from a specific store and domain
  jenkins credential delete my-cred-id --store system --domain my-domain --confirm`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !confirm {
				return fmt.Errorf("use --confirm to confirm deletion of credential %q", id)
			}

			if err := jenkinsClient.DeleteCredential(store, domain, id); err != nil {
				return fmt.Errorf("deleting credential: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Credential %q deleted.\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&store, "store", "system", "Credential store")
	cmd.Flags().StringVar(&domain, "domain", "_", "Credential domain")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm deletion")

	return cmd
}
