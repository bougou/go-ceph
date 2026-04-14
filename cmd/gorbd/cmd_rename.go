package main

import (
	"context"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rename <src-image-spec> <dst-image-spec>",
		Aliases: []string{"mv"},
		Short:   "Rename image",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				return conn.RbdRename(context.Background(), ceph.ImageSpec(args[0]), ceph.ImageSpec(args[1]))
			})
		},
	}
}
