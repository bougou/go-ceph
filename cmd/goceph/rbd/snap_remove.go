package rbd

import (
	"context"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newSnapRemoveCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:     "rm <snap-spec>",
		Aliases: []string{"remove"},
		Short:   "Remove snapshot",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdSnapRemove(context.Background(), rbdapi.SnapSpec(args[0]))
			})
		},
	}
}
