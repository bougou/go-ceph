package main

import (
	"context"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newCopyCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "copy <src-image-spec> <dst-image-spec>",
		Aliases: []string{"cp"},
		Short:   "Copy image",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				return conn.RbdCopy(context.Background(), ceph.ImageSpec(args[0]), ceph.ImageSpec(args[1]))
			})
		},
	}
}
