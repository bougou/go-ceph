package main

import (
	"context"
	"fmt"

	"github.com/bougou/go-ceph/pkg/rados"
	"github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <image-spec>",
		Short: "Show image info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *rados.RadosConn) error {
				info, err := conn.RbdInfo(context.Background(), rbd.ImageSpec(args[0]))
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
