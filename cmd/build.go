package cmd

import (
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "build",
		Aliases: []string{"builds"},
		Short:   "Manage Jenkins builds",
		Long:    "List, inspect, and manage Jenkins build runs.",
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
