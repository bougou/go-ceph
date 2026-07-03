package ceph

import (
	"github.com/bougou/go-ceph/cmd/goceph/ceph/rbd"
	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/spf13/cobra"
)

// NewCmd builds the goceph ceph command group.
func NewCmd(opts *app.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ceph",
		Short: "Ceph cluster operations",
		Long:  "Commands that correspond to the native `ceph` CLI and mgr modules.",
	}

	cmd.AddCommand(rbd.NewCmd(opts))

	return cmd
}
