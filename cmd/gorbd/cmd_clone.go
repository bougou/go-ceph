package main

import (
	"context"

	"github.com/bougou/go-ceph/pkg/rados"
	"github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newCloneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clone <src-snap-spec> <dst-image-spec>",
		Short: "Clone image from snapshot",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdClone(context.Background(), rbd.SnapSpec(args[0]), rbd.ImageSpec(args[1]))
			})
		},
	}
}
