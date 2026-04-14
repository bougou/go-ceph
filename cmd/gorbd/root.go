package main

import "github.com/spf13/cobra"

type globalOptions struct {
	cephConf string
	retries  int
}

var globalOpts globalOptions

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gorbd",
		Short: "RBD helper CLI based on go-ceph",
	}

	cmd.PersistentFlags().StringVarP(&globalOpts.cephConf, "conf", "c", "", "Ceph config path")
	cmd.PersistentFlags().IntVarP(&globalOpts.retries, "retries", "r", 0, "Retry count for operations")

	cmd.AddCommand(
		newCreateCmd(),
		newInfoCmd(),
		newRemoveCmd(),
		newRenameCmd(),
		newResizeCmd(),
		newFlattenCmd(),
		newCopyCmd(),
		newCloneCmd(),
		newChildrenCmd(),
		newSnapCmd(),
	)

	return cmd
}
