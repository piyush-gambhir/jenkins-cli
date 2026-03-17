package cmd

import (
	"github.com/spf13/cobra"
)

func newJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "job",
		Aliases: []string{"jobs"},
		Short:   "Manage Jenkins jobs",
		Long:    "List, create, update, and manage Jenkins jobs and their configurations.",
	}

	cmd.AddCommand(newJobListCmd())
	cmd.AddCommand(newJobGetCmd())
	cmd.AddCommand(newJobCreateCmd())
	cmd.AddCommand(newJobUpdateCmd())
	cmd.AddCommand(newJobCopyCmd())
	cmd.AddCommand(newJobRenameCmd())
	cmd.AddCommand(newJobDeleteCmd())
	cmd.AddCommand(newJobEnableCmd())
	cmd.AddCommand(newJobDisableCmd())
	cmd.AddCommand(newJobConfigCmd())
	cmd.AddCommand(newJobWipeWorkspaceCmd())
	cmd.AddCommand(newJobBuildCmd())

	return cmd
}
