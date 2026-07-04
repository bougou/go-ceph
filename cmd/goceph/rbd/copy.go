package rbd

import (
	"context"
	"strings"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newCopyCmd(opts *app.Options) *cobra.Command {
	var unsafe bool

	cmd := &cobra.Command{
		Use:     "copy <src-image-or-snap-spec> <dst-image-spec>",
		Aliases: []string{"cp"},
		Short:   "Copy image (auto snapshot for consistency; use --unsafe to copy head directly)",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				dstImageSpec := rbdapi.ImageSpec(args[1])
				if unsafe {
					return conn.RbdCopyUnsafe(context.Background(), rbdapi.ImageSpec(args[0]), dstImageSpec)
				}

				if strings.Contains(args[0], "@") {
					return conn.RbdCopySnap(context.Background(), rbdapi.SnapSpec(args[0]), dstImageSpec)
				}
				return conn.RbdCopy(context.Background(), rbdapi.ImageSpec(args[0]), dstImageSpec)
			})
		},
	}

	cmd.Flags().BoolVar(&unsafe, "unsafe", false, "Copy image head directly without creating a snapshot (may be inconsistent)")

	return cmd
}
