package cmd

import (
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "build",
		Aliases: []string{"builds"},
		Short:   "Manage Jenkins builds",
		Long: `List, inspect, and manage Jenkins build runs.

Subcommands:
  list          List builds for a job (with optional status filter)
  get           Get detailed information about a specific build
  log           View or stream build console output
  stop          Stop a running build
  delete        Delete a build record
  artifacts     List or download build artifacts
  test-report   View test results for a build
  env           View injected environment variables
  stages        View pipeline stage breakdown
  replay        Replay a pipeline build
  open          Open a build in the browser`,
	}

	cmd.AddCommand(newBuildListCmd())
	cmd.AddCommand(newBuildGetCmd())
	cmd.AddCommand(newBuildLogCmd())
	cmd.AddCommand(newBuildStopCmd())
	cmd.AddCommand(newBuildDeleteCmd())
	cmd.AddCommand(newBuildArtifactsCmd())
	cmd.AddCommand(newBuildTestReportCmd())
	cmd.AddCommand(newBuildEnvCmd())
	cmd.AddCommand(newBuildStagesCmd())
	cmd.AddCommand(newBuildReplayCmd())
	cmd.AddCommand(newBuildOpenCmd())

	return cmd
}
