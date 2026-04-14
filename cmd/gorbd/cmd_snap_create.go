package main

import (
	"context"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newSnapCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <snap-spec>",
		Short: "Create snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				return conn.RbdSnapCreate(context.Background(), ceph.SnapSpec(args[0]))
			})
		},
	}
}
