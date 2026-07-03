package task

import (
	"context"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

// NewCmd builds the goceph ceph rbd task command group.
func NewCmd(opts *app.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage asynchronous RBD background tasks",
		Long:  "Mgr-backed RBD tasks (submitted via ceph mgr, not librbd).",
	}
	cmd.AddCommand(
		newAddCmd(opts),
		newListCmd(opts),
		newGetCmd(opts),
		newCancelCmd(opts),
	)
	return cmd
}

func newAddCmd(opts *app.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a background task",
	}
	cmd.AddCommand(
		newAddFlattenCmd(opts),
		newAddRemoveCmd(opts),
		newAddTrashRemoveCmd(opts),
	)
	return cmd
}

func newAddFlattenCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "flatten <image-spec>",
		Short: "Add a background flatten task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				task, err := conn.RbdTaskAddFlatten(context.Background(), rbdapi.ImageSpec(args[0]))
				if err != nil {
					return err
				}
				return app.WriteJSON(cmd.OutOrStdout(), task, true)
			})
		},
	}
}

func newAddRemoveCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <image-spec>",
		Short: "Add a background remove task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				task, err := conn.RbdTaskAddRemove(context.Background(), rbdapi.ImageSpec(args[0]))
				if err != nil {
					return err
				}
				return app.WriteJSON(cmd.OutOrStdout(), task, true)
			})
		},
	}
}

func newAddTrashRemoveCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "trash-remove <image-id-spec>",
		Short: "Add a background trash remove task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				task, err := conn.RbdTaskAddTrashRemove(context.Background(), rbdapi.ImageSpec(args[0]))
				if err != nil {
					return err
				}
				return app.WriteJSON(cmd.OutOrStdout(), task, true)
			})
		},
	}
}

func newListCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List background tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				tasks, err := conn.RbdTaskList(context.Background())
				if err != nil {
					return err
				}
				return app.WriteJSON(cmd.OutOrStdout(), tasks, true)
			})
		},
	}
}

func newGetCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <task-id>",
		Short: "Get a background task by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				task, err := conn.RbdTaskGet(context.Background(), args[0])
				if err != nil {
					return err
				}
				return app.WriteJSON(cmd.OutOrStdout(), task, true)
			})
		},
	}
}

func newCancelCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <task-id>",
		Short: "Cancel a background task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				task, err := conn.RbdTaskCancel(context.Background(), args[0])
				if err != nil {
					return err
				}
				return app.WriteJSON(cmd.OutOrStdout(), task, true)
			})
		},
	}
}
