package main

import (
	"context"

	"github.com/bougou/go-ceph/pkg/rados"
	"github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rename <src-image-spec> <dst-image-spec>",
		Aliases: []string{"mv"},
		Short:   "Rename image",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdRename(context.Background(), rbd.ImageSpec(args[0]), rbd.ImageSpec(args[1]))
			})
		},
	}
}
