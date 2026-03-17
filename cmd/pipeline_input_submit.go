package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newPipelineInputSubmitCmd() *cobra.Command {
	var params []string

	cmd := &cobra.Command{
		Use:   "input-submit <job-path> <build-number> <input-id>",
		Short: "Submit a pipeline input",
		Long: `Proceed with a pending pipeline input action, optionally providing parameters.

Use "jenkins pipeline input-list" to find the input ID, then submit it
with this command. Use --param to provide input parameters if required.

Examples:
  # Submit an input (proceed)
  jenkins pipeline input-submit my-pipeline 42 my-input-id

  # Submit with parameters
  jenkins pipeline input-submit my-pipeline 42 my-input-id --param APPROVE=yes

  # Submit with multiple parameters
  jenkins pipeline input-submit my-pipeline 42 deploy-approval --param ENV=prod --param VERSION=1.0`,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]
			number, err := parseNumber(args[1])
			if err != nil {
				return err
			}
			inputID := args[2]

			paramMap := make(map[string]string)
			for _, p := range params {
				parts := strings.SplitN(p, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid parameter format %q, expected KEY=VALUE", p)
				}
				paramMap[parts[0]] = parts[1]
			}

			if err := jenkinsClient.SubmitPipelineInput(jobPath, number, inputID, paramMap); err != nil {
				return fmt.Errorf("submitting input: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Input %q submitted for build #%d.\n", inputID, number)
			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&params, "param", "p", nil, "Input parameters (KEY=VALUE, repeatable)")

	return cmd
}
