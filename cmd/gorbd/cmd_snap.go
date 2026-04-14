package main

import "github.com/spf13/cobra"

func newSnapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snap",
		Short: "Snapshot operations",
	}
	cmd.AddCommand(
		newSnapListCmd(),
		newSnapCreateCmd(),
		newSnapRemoveCmd(),
	)
	return cmd
}
