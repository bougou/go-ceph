package main

import (
	"context"

	"github.com/bougou/go-ceph/pkg/rados"
	"github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newCopyCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "copy <src-image-spec> <dst-image-spec>",
		Aliases: []string{"cp"},
		Short:   "Copy image",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdCopy(context.Background(), rbd.ImageSpec(args[0]), rbd.ImageSpec(args[1]))
			})
		},
	}
}
