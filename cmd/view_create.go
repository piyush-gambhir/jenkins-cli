package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newViewCreateCmd() *cobra.Command {
	var viewType string

	cmd := &cobra.Command{
		Use:   "create <view-name>",
		Short: "Create a new view",
		Long:  "Create a new Jenkins view.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := jenkinsClient.CreateView(name, viewType); err != nil {
				return fmt.Errorf("creating view: %w", err)
			}

			fmt.Fprintf(os.Stdout, "View %q created.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&viewType, "type", "hudson.model.ListView", "View type class name")

	return cmd
}
