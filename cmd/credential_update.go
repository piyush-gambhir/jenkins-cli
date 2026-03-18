package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func newCredentialUpdateCmd() *cobra.Command {
	var store string
	var domain string
	var fromFile string

	cmd := &cobra.Command{
		Use:         "update <credential-id>",
		Short:       "Update a credential",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Update an existing credential from an XML configuration file.

The --from-file flag is required. This replaces the credential's entire
configuration with the contents of the XML file. Use "-" as the file
path to read from stdin.

Examples:
  # Update a credential
  jenkins credential update my-cred-id --from-file updated-cred.xml

  # Update from stdin
  cat cred.xml | jenkins credential update my-cred-id --from-file -

  # Update in a specific store and domain
  jenkins credential update my-cred-id --from-file cred.xml --store system --domain my-domain`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

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

			if err := jenkinsClient.UpdateCredential(store, domain, id, string(data)); err != nil {
				return fmt.Errorf("updating credential: %w", err)
			}

			if !quietFlag {
				fmt.Fprintf(os.Stdout, "Credential %q updated successfully.\n", id)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&store, "store", "system", "Credential store")
	cmd.Flags().StringVar(&domain, "domain", "_", "Credential domain")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to XML config file (required, use - for stdin)")

	return cmd
}
