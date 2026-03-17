package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCredentialUpdateCmd() *cobra.Command {
	var store string
	var domain string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "update <credential-id>",
		Short: "Update a credential",
		Long:  "Update an existing credential from an XML configuration file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if fromFile == "" {
				return fmt.Errorf("--from-file is required")
			}

			data, err := os.ReadFile(fromFile)
			if err != nil {
				return fmt.Errorf("reading config file %s: %w", fromFile, err)
			}

			if err := jenkinsClient.UpdateCredential(store, domain, id, string(data)); err != nil {
				return fmt.Errorf("updating credential: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Credential %q updated successfully.\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&store, "store", "system", "Credential store")
	cmd.Flags().StringVar(&domain, "domain", "_", "Credential domain")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to XML config file (required)")

	return cmd
}
