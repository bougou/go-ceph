package rbd

import (
	"context"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newSnapCreateCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "create <snap-spec>",
		Short: "Create snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdSnapCreate(context.Background(), rbdapi.SnapSpec(args[0]))
			})
		},
	}
}
