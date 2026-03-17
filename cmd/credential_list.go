package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newCredentialListCmd() *cobra.Command {
	var store string
	var domain string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List credentials",
		Long:  "List all credentials in a store and domain.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := jenkinsClient.ListCredentials(store, domain)
			if err != nil {
				return fmt.Errorf("listing credentials: %w", err)
			}

			if len(creds) == 0 {
				fmt.Fprintln(os.Stdout, "No credentials found.")
				return nil
			}

			tableDef := &output.TableDef{
				Headers: []string{"ID", "TYPE", "DISPLAY NAME", "DESCRIPTION"},
				RowFunc: func(item interface{}) []string {
					c := item.(client.Credential)
					desc := c.Description
					if len(desc) > 50 {
						desc = desc[:50] + "..."
					}
					return []string{c.ID, c.TypeName, c.DisplayName, desc}
				},
			}

			return output.Print(os.Stdout, outFormat, creds, tableDef)
		},
	}

	cmd.Flags().StringVar(&store, "store", "system", "Credential store")
	cmd.Flags().StringVar(&domain, "domain", "_", "Credential domain")

	return cmd
}
