package main

import (
	"context"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newSnapRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm <snap-spec>",
		Aliases: []string{"remove"},
		Short:   "Remove snapshot",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				return conn.RbdSnapRemove(context.Background(), ceph.SnapSpec(args[0]))
			})
		},
	}
}
