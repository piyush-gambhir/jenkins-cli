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
		Use:         "create",
		Short:       "Create a credential",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Create a new credential from an XML configuration file.

The --from-file flag is required and must point to a valid Jenkins
credentials XML file. Use --store and --domain to target a specific
credential store and domain.

Example XML for a username/password credential:
  <com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl>
    <scope>GLOBAL</scope>
    <id>my-cred-id</id>
    <username>admin</username>
    <password>secret</password>
    <description>My credential</description>
  </com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl>

Examples:
  # Create a credential from XML
  jenkins credential create --from-file cred.xml

  # Create in a specific store and domain
  jenkins credential create --from-file cred.xml --store system --domain my-domain`,
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
