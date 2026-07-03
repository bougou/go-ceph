package rbd

import (
	"context"
	"fmt"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newChildrenCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "children <image-or-snap-spec>",
		Short: "Display children of an image or snapshot",
		Long: `Display children of an image or its snapshot.

Positional arguments:
  <image-or-snap-spec>  image or snapshot specification
                        ([<pool-name>/[<namespace>/]]<image-name>[@<snap-name>])`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				spec := args[0]
				_, _, _, snapshotName, err := rbdapi.ImageOrSnap(spec)
				if err != nil {
					return err
				}

				var (
					children []rbdapi.ImageSpec
				)

				if snapshotName != "" {
					children, err = conn.RbdSnapChildren(context.Background(), rbdapi.SnapSpec(spec))
				} else {
					children, err = conn.RbdChildren(context.Background(), rbdapi.ImageSpec(spec))
				}
				if err != nil {
					return err
				}
				for _, child := range children {
					fmt.Println(string(child))
				}
				return nil
			})
		},
	}
}
