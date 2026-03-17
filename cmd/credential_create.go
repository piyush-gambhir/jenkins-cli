package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCredentialCreateCmd() *cobra.Command {
	var store string
	var domain string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a credential",
		Long:  "Create a new credential from an XML configuration file.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if fromFile == "" {
				return fmt.Errorf("--from-file is required")
			}

			data, err := os.ReadFile(fromFile)
			if err != nil {
				return fmt.Errorf("reading config file %s: %w", fromFile, err)
			}

			if err := jenkinsClient.CreateCredential(store, domain, string(data)); err != nil {
				return fmt.Errorf("creating credential: %w", err)
			}

			fmt.Fprintln(os.Stdout, "Credential created successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&store, "store", "system", "Credential store")
	cmd.Flags().StringVar(&domain, "domain", "_", "Credential domain")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to XML config file (required)")

	return cmd
}
