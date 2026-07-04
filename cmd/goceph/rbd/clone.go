package rbd

import (
	"context"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newCloneCmd(opts *app.Options) *cobra.Command {
	var flattenMode string

	cmd := &cobra.Command{
		Use:   "clone <src-snap-spec> <dst-image-spec>",
		Short: "Clone image from snapshot",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				flattenOpt, err := rbdapi.ParseCloneAutoFlattenOption(flattenMode)
				if err != nil {
					return err
				}

				var cloneOpts []rbdapi.CloneOption
				if flattenOpt != nil {
					cloneOpts = append(cloneOpts, flattenOpt)
				}

				task, err := conn.RbdClone(context.Background(), rbdapi.SnapSpec(args[0]), rbdapi.ImageSpec(args[1]), cloneOpts...)
				if err != nil {
					return err
				}
				if task != nil {
					return app.WriteJSON(cmd.OutOrStdout(), task, true)
				}
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&flattenMode, "flatten", "", "Auto-flatten: none (default), or depth threshold 0..15")

	return cmd
}
