package rados

import (
	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/spf13/cobra"
)

// NewCmd builds the goceph rados command group.
func NewCmd(opts *app.Options) *cobra.Command {
	_ = opts
	return &cobra.Command{
		Use:   "rados",
		Short: "RADOS operations",
		Long:  "Commands that correspond to the native `rados` CLI.",
	}
}
