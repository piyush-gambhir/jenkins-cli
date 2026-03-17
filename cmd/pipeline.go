package cmd

import (
	"github.com/spf13/cobra"
)

func newPipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Pipeline operations",
		Long:  "Validate Jenkinsfiles and manage pipeline input actions.",
	}

	cmd.AddCommand(newPipelineValidateCmd())
	cmd.AddCommand(newPipelineInputListCmd())
	cmd.AddCommand(newPipelineInputSubmitCmd())
	cmd.AddCommand(newPipelineInputAbortCmd())

	return cmd
}
