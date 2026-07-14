package cmd

import (
	"github.com/spf13/cobra"
)

func newJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "job",
		Aliases: []string{"jobs"},
		Short:   "Manage Jenkins jobs",
		Long: `List, create, update, and manage Jenkins jobs and their configurations.

Jenkins organizes jobs in folders using slash-separated paths:
  my-job                      root-level job
  my-folder/my-job            job in a folder
  team/project/pipeline       nested folders

Subcommands:
  list             List jobs (with optional folder, recursive, and status filters)
  get              Get detailed information about a job
  create           Create a new job from XML config
  update           Update a job's XML config
  copy             Copy an existing job
  rename           Rename a job
  delete           Permanently delete a job
  enable           Enable a disabled job
  disable          Disable a job
  config           Retrieve the raw config.xml
  wipe-workspace   Wipe the workspace directory
  build            Trigger a build (with optional params, wait, follow)`,
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
