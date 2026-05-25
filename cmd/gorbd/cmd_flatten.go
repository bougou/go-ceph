package main

import (
	"context"

	"github.com/bougou/go-ceph/pkg/rados"
	"github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newFlattenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "flatten <image-spec>",
		Short: "Flatten image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdFlatten(context.Background(), rbd.ImageSpec(args[0]))
			})
		},
	}
}
