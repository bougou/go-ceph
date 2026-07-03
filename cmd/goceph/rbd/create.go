package rbd

import (
	"context"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newCreateCmd(opts *app.Options) *cobra.Command {
	var sizeStr string

	cmd := &cobra.Command{
		Use:   "create [--size <size>] <image-spec>",
		Short: "Create image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sizeBytes, err := app.ParseSizeToBytes(sizeStr)
			if err != nil {
				return err
			}
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdCreate(context.Background(), rbdapi.ImageSpec(args[0]), int64(sizeBytes))
			})
		},
	}
	cmd.Flags().StringVarP(&sizeStr, "size", "s", "", "Image size, e.g. 1G, 1024M, 1073741824")
	_ = cmd.MarkFlagRequired("size")
	return cmd
}
