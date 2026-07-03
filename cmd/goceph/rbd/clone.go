package rbd

import (
	"context"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newCloneCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "clone <src-snap-spec> <dst-image-spec>",
		Short: "Clone image from snapshot",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdClone(context.Background(), rbdapi.SnapSpec(args[0]), rbdapi.ImageSpec(args[1]))
			})
		},
	}
}
