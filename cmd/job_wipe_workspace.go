package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newJobWipeWorkspaceCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "wipe-workspace <job-path>",
		Short: "Wipe job workspace",
		Long:  "Wipe the workspace directory of a Jenkins job.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobPath := args[0]

			if !confirm {
				return fmt.Errorf("use --confirm to confirm wiping workspace for job %q", jobPath)
			}

			if err := jenkinsClient.WipeWorkspace(jobPath); err != nil {
				return fmt.Errorf("wiping workspace: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Workspace for job %q wiped.\n", jobPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm wipe")

	return cmd
}
