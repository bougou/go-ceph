package main

import (
	"context"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newFlattenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "flatten <image-spec>",
		Short: "Flatten image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				return conn.RbdFlatten(context.Background(), ceph.ImageSpec(args[0]))
			})
		},
	}
}
