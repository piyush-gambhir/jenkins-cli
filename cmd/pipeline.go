package cmd

import (
	"github.com/spf13/cobra"
)

func newPipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Pipeline operations",
		Long: `Validate Jenkinsfiles and manage pipeline input actions.

Subcommands:
  validate       Validate a declarative Jenkinsfile
  input-list     List pending input actions for a pipeline build
  input-submit   Proceed with a pending input action
  input-abort    Abort a pending input action`,
	}

	cmd.AddCommand(newPipelineValidateCmd())
	cmd.AddCommand(newPipelineInputListCmd())
	cmd.AddCommand(newPipelineInputSubmitCmd())
	cmd.AddCommand(newPipelineInputAbortCmd())

	return cmd
}
