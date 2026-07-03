package rbd

import (
	"context"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newRenameCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:     "rename <src-image-spec> <dst-image-spec>",
		Aliases: []string{"mv"},
		Short:   "Rename image",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdRename(context.Background(), rbdapi.ImageSpec(args[0]), rbdapi.ImageSpec(args[1]))
			})
		},
	}
}
