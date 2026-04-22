package main

import (
	"context"
	"fmt"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newChildrenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "children <image-or-snap-spec>",
		Short: "Display children of an image or snapshot",
		Long: `Display children of an image or its snapshot.

Positional arguments:
  <image-or-snap-spec>  image or snapshot specification
                        ([<pool-name>/[<namespace>/]]<image-name>[@<snap-name>])`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				spec := args[0]
				_, _, _, snapshotName, err := ceph.ImageOrSnap(spec)
				if err != nil {
					return err
				}

				var (
					children []ceph.ImageSpec
				)

				if snapshotName != "" {
					children, err = conn.RbdSnapChildren(context.Background(), ceph.SnapSpec(spec))
				} else {
					children, err = conn.RbdChildren(context.Background(), ceph.ImageSpec(spec))
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
