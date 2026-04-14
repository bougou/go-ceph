package main

import (
	"context"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	var sizeStr string

	cmd := &cobra.Command{
		Use:   "create [--size <size>] <image-spec>",
		Short: "Create image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sizeBytes, err := parseSizeToBytes(sizeStr)
			if err != nil {
				return err
			}
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				return conn.RbdCreate(context.Background(), ceph.ImageSpec(args[0]), int64(sizeBytes))
			})
		},
	}
	cmd.Flags().StringVarP(&sizeStr, "size", "s", "", "Image size, e.g. 1G, 1024M, 1073741824")
	_ = cmd.MarkFlagRequired("size")
	return cmd
}
