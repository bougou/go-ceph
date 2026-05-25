package main

import (
	"context"

	"github.com/bougou/go-ceph/pkg/rados"
	"github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newSnapCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <snap-spec>",
		Short: "Create snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdSnapCreate(context.Background(), rbd.SnapSpec(args[0]))
			})
		},
	}
}
