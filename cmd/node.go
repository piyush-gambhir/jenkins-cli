package cmd

import (
	"github.com/spf13/cobra"
)

func newNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "node",
		Aliases: []string{"nodes", "agent", "agents"},
		Short:   "Manage Jenkins nodes/agents",
		Long:    "List, create, and manage Jenkins build nodes and agents.",
	}

	cmd.AddCommand(newNodeListCmd())
	cmd.AddCommand(newNodeGetCmd())
	cmd.AddCommand(newNodeCreateCmd())
	cmd.AddCommand(newNodeDeleteCmd())
	cmd.AddCommand(newNodeEnableCmd())
	cmd.AddCommand(newNodeDisableCmd())
	cmd.AddCommand(newNodeLogCmd())

	return cmd
}
