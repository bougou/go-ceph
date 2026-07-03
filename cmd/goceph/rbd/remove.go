package rbd

import (
	"context"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newRemoveCmd(opts *app.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <image-spec>",
		Aliases: []string{"remove"},
		Short:   "Remove image",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdRemove(context.Background(), rbdapi.ImageSpec(args[0]))
			})
		},
	}
	return cmd
}
