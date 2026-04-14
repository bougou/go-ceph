package main

import (
	"context"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newCloneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clone <src-snap-spec> <dst-image-spec>",
		Short: "Clone image from snapshot",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				return conn.RbdClone(context.Background(), ceph.SnapSpec(args[0]), ceph.ImageSpec(args[1]))
			})
		},
	}
}
