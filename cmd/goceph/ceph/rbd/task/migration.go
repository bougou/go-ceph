package task

import (
	"context"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newAddMigrationCmd(opts *app.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migration",
		Short: "Add background migration tasks",
	}
	cmd.AddCommand(
		newAddMigrationExecuteCmd(opts),
		newAddMigrationCommitCmd(opts),
		newAddMigrationAbortCmd(opts),
	)
	return cmd
}

func newAddMigrationExecuteCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "execute <image-spec>",
		Short: "Add a background migration execute task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				task, err := conn.RbdTaskAddMigrationExecute(context.Background(), rbdapi.ImageSpec(args[0]))
				if err != nil {
					return err
				}
				return app.WriteJSON(cmd.OutOrStdout(), task, true)
			})
		},
	}
}

func newAddMigrationCommitCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "commit <image-spec>",
		Short: "Add a background migration commit task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				task, err := conn.RbdTaskAddMigrationCommit(context.Background(), rbdapi.ImageSpec(args[0]))
				if err != nil {
					return err
				}
				return app.WriteJSON(cmd.OutOrStdout(), task, true)
			})
		},
	}
}

func newAddMigrationAbortCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "abort <image-spec>",
		Short: "Add a background migration abort task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				task, err := conn.RbdTaskAddMigrationAbort(context.Background(), rbdapi.ImageSpec(args[0]))
				if err != nil {
					return err
				}
				return app.WriteJSON(cmd.OutOrStdout(), task, true)
			})
		},
	}
}
