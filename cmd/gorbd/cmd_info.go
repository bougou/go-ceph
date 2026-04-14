package main

import (
	"context"
	"fmt"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <image-spec>",
		Short: "Show image info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				info, err := conn.RbdInfo(context.Background(), ceph.ImageSpec(args[0]))
				if err != nil {
					return err
				}
				if info == nil {
					return fmt.Errorf("image %q not found", args[0])
				}
				fmt.Println(info.String())
				return nil
			})
		},
	}
}
