package cmd

import (
	"github.com/spf13/cobra"
)

func newNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "node",
		Aliases: []string{"nodes", "agent", "agents"},
		Short:   "Manage Jenkins nodes/agents",
		Long: `List, create, and manage Jenkins build nodes and agents.

Subcommands:
  list      List all nodes (with optional --offline / --online filters)
  get       Get detailed information about a node
  create    Create a new permanent agent node
  delete    Delete a node
  enable    Bring an offline node back online
  disable   Take a node offline (with optional message)
  log       View the agent log for a node`,
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
