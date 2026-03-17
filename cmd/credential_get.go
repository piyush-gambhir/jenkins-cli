package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newCredentialGetCmd() *cobra.Command {
	var store string
	var domain string

	cmd := &cobra.Command{
		Use:   "get <credential-id>",
		Short: "Get credential details",
		Long:  "Display details about a specific credential.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			cred, err := jenkinsClient.GetCredential(store, domain, id)
			if err != nil {
				return fmt.Errorf("getting credential: %w", err)
			}

			if outFormat == output.FormatTable {
				fmt.Fprintf(os.Stdout, "Credential: %s\n", cred.ID)
				fmt.Fprintf(os.Stdout, "  Type:        %s\n", cred.TypeName)
				fmt.Fprintf(os.Stdout, "  Display:     %s\n", cred.DisplayName)
				if cred.Description != "" {
					fmt.Fprintf(os.Stdout, "  Description: %s\n", cred.Description)
				}
				return nil
			}

			return output.Print(os.Stdout, outFormat, cred, nil)
		},
	}

	cmd.Flags().StringVar(&store, "store", "system", "Credential store")
	cmd.Flags().StringVar(&domain, "domain", "_", "Credential domain")

	return cmd
}
