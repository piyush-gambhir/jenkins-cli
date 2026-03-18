package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
)

func newCredentialCreateCmd() *cobra.Command {
	var store string
	var domain string
	var fromFile string
	var ifNotExists bool

	cmd := &cobra.Command{
		Use:         "create",
		Short:       "Create a credential",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Create a new credential from an XML configuration file.

The --from-file flag is required and must point to a valid Jenkins
credentials XML file. Use --store and --domain to target a specific
credential store and domain. Use "-" as the file path to read from stdin.

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

  # Create from stdin
  cat cred.xml | jenkins credential create --from-file -

  # Idempotent create (no error if credential already exists)
  jenkins credential create --from-file cred.xml --if-not-exists`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if fromFile == "" {
				return fmt.Errorf("--from-file is required")
			}

			var data []byte
			var err error
			if fromFile == "-" {
				data, err = io.ReadAll(os.Stdin)
			} else {
				data, err = os.ReadFile(fromFile)
			}
			if err != nil {
				return fmt.Errorf("reading config file %s: %w", fromFile, err)
			}

			if err := jenkinsClient.CreateCredential(store, domain, string(data)); err != nil {
				var apiErr *client.APIError
				if ifNotExists && errors.As(err, &apiErr) && (apiErr.StatusCode == 400 || apiErr.StatusCode == 409) {
					if !quietFlag {
						fmt.Fprintln(os.Stdout, "Credential already exists, skipping.")
					}
					return nil
				}
				return fmt.Errorf("creating credential: %w", err)
			}

			if !quietFlag {
				fmt.Fprintln(os.Stdout, "Credential created successfully.")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&store, "store", "system", "Credential store")
	cmd.Flags().StringVar(&domain, "domain", "_", "Credential domain")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to XML config file (required, use - for stdin)")
	cmd.Flags().BoolVar(&ifNotExists, "if-not-exists", false, "Don't error if the credential already exists")

	return cmd
}
