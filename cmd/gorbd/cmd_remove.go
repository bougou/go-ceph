package main

import (
	"context"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <image-spec>",
		Aliases: []string{"remove"},
		Short:   "Remove image",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				return conn.RbdRemove(context.Background(), ceph.ImageSpec(args[0]))
			})
		},
	}
	return cmd
}
