package rbd

import (
	"context"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newResizeCmd(opts *app.Options) *cobra.Command {
	var sizeStr string

	cmd := &cobra.Command{
		Use:   "resize [--size <size>] <image-spec>",
		Short: "Resize image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sizeBytes, err := app.ParseSizeToBytes(sizeStr)
			if err != nil {
				return err
			}
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdResize(context.Background(), rbdapi.ImageSpec(args[0]), sizeBytes)
			})
		},
	}
	cmd.Flags().StringVarP(&sizeStr, "size", "s", "", "New size, e.g. 2G")
	_ = cmd.MarkFlagRequired("size")
	return cmd
}
