package main

import (
	"context"
	"fmt"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newSnapListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ls <image-spec>",
		Aliases: []string{"list"},
		Short:   "List snapshots",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				snaps, err := conn.RbdSnapList(context.Background(), ceph.ImageSpec(args[0]))
				if err != nil {
					return err
				}
				for _, s := range snaps {
					fmt.Println(s)
				}
				return nil
			})
		},
	}
}
