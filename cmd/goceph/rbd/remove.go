package rbd

import (
	"context"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newRemoveCmd(opts *app.Options) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "rm <image-spec>",
		Aliases: []string{"remove"},
		Short:   "Remove image",
		Long: `Remove an RBD image.

With --force, clone children are flattened (or permanently deleted if already
in trash), all snapshots are unprotected and removed, then the image itself
is deleted. Live child images are preserved as independent images.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				imageSpec := rbdapi.ImageSpec(args[0])
				if force {
					return conn.RbdForceRemove(context.Background(), imageSpec)
				}
				return conn.RbdRemove(context.Background(), imageSpec)
			})
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "flatten clone children, remove all snapshots, then remove the image")
	return cmd
}
