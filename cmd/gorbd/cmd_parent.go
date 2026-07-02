package main

import (
	"context"
	"fmt"

	"github.com/bougou/go-ceph/pkg/rados"
	"github.com/bougou/go-ceph/pkg/rbd"
	cephrbd "github.com/ceph/go-ceph/rbd"
	"github.com/spf13/cobra"
)

func newParentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "parent <image-or-snap-spec>",
		Short: "Display parent images of an image or snapshot",
		Long: `Display parent images of an image or snapshot, walking up the clone chain.

The nearest parent is listed first.

Positional arguments:
  <image-or-snap-spec>  image or snapshot specification
                        ([<pool-name>/[<namespace>/]]<image-name>[@<snap-name>])`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *rados.RadosConn) error {
				spec := args[0]
				_, _, _, snapshotName, err := rbd.ImageOrSnap(spec)
				if err != nil {
					return err
				}

				var parents []*cephrbd.ParentInfo

				if snapshotName != "" {
					parents, err = conn.RbdSnapParents(context.Background(), rbd.SnapSpec(spec))
				} else {
					parents, err = conn.RbdParents(context.Background(), rbd.ImageSpec(spec))
				}
				if err != nil {
					return err
				}
				for _, parent := range parents {
					fmt.Printf("%s/%s@%s\n", parent.Image.PoolName, parent.Image.ImageName, parent.Snap.SnapName)
				}
				return nil
			})
		},
	}
}
