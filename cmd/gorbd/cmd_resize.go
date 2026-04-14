package main

import (
	"context"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newResizeCmd() *cobra.Command {
	var sizeStr string

	cmd := &cobra.Command{
		Use:   "resize [--size <size>] <image-spec>",
		Short: "Resize image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sizeBytes, err := parseSizeToBytes(sizeStr)
			if err != nil {
				return err
			}
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				return conn.RbdResize(context.Background(), ceph.ImageSpec(args[0]), sizeBytes)
			})
		},
	}
	cmd.Flags().StringVarP(&sizeStr, "size", "s", "", "New size, e.g. 2G")
	_ = cmd.MarkFlagRequired("size")
	return cmd
}
