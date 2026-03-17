package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/output"
)

func newCredentialListCmd() *cobra.Command {
	var store string
	var domain string
	var credType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List credentials",
		Long: `List all credentials in a store and domain.

By default lists credentials from the "system" store in the global ("_")
domain. Use --store and --domain to target a different store/domain. Use
--type to filter by credential type name (case-insensitive substring match).

Jenkins credential stores:
  system   - System-level credentials (default)
  folder   - Folder-level credentials (requires folder context)

Common credential types:
  Username with password
  SSH Username with private key
  Secret text
  Secret file
  Certificate

Examples:
  # List all system credentials
  jenkins credential list

  # List credentials in a specific store and domain
  jenkins credential list --store system --domain my-domain

  # Filter credentials by type (substring match)
  jenkins credential list --type "SSH"

  # List credentials as JSON
  jenkins credential list -o json

  # List only username/password credentials
  jenkins credential list --type "Username with password"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := jenkinsClient.ListCredentials(store, domain)
			if err != nil {
				return fmt.Errorf("listing credentials: %w", err)
			}

			// Filter by type if specified
			if credType != "" {
				typeLower := strings.ToLower(credType)
				var filtered []client.Credential
				for _, c := range creds {
					if strings.Contains(strings.ToLower(c.TypeName), typeLower) {
						filtered = append(filtered, c)
					}
				}
				creds = filtered
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

	cmd.Flags().StringVar(&store, "store", "system", "Credential store (e.g. system, folder)")
	cmd.Flags().StringVar(&domain, "domain", "_", "Credential domain (use _ for global domain)")
	cmd.Flags().StringVar(&credType, "type", "", "Filter by credential type name (case-insensitive substring match)")

	return cmd
}
