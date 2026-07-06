package rbd

import (
	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/spf13/cobra"
)

// NewCmd builds the goceph rbd command group.
func NewCmd(opts *app.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rbd",
		Short: "RBD image operations",
		Long:  "Commands that correspond to the native `rbd` CLI.",
	}

	cmd.AddCommand(
		newCreateCmd(opts),
		newListCmd(opts),
		newInfoCmd(opts),
		newStatusCmd(opts),
		newRemoveCmd(opts),
		newRenameCmd(opts),
		newResizeCmd(opts),
		newFlattenCmd(opts),
		newCopyCmd(opts),
		newCloneCmd(opts),
		newChildrenCmd(opts),
		newParentCmd(opts),
		newMigrationCmd(opts),
		newSnapCmd(opts),
		newDeviceCmd(opts),
	)

	return cmd
}
