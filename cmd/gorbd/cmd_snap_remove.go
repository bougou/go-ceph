package main

import (
	"context"

	"github.com/bougou/go-ceph/pkg/rados"
	"github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newSnapRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm <snap-spec>",
		Aliases: []string{"remove"},
		Short:   "Remove snapshot",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdSnapRemove(context.Background(), rbd.SnapSpec(args[0]))
			})
		},
	}
}
