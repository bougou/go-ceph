package rbd

import (
	"context"
	"fmt"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newMigrationCmd(opts *app.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migration",
		Short: "Manage RBD image live migration",
	}
	cmd.AddCommand(
		newMigrationPrepareCmd(opts),
		newMigrationExecuteCmd(opts),
		newMigrationCommitCmd(opts),
		newMigrationAbortCmd(opts),
		newMigrationStatusCmd(opts),
	)
	return cmd
}

func newMigrationPrepareCmd(opts *app.Options) *cobra.Command {
	var importOnly bool
	var sourceSpec string

	cmd := &cobra.Command{
		Use:   "prepare <source-image-spec> [<dest-image-spec>]",
		Short: "Prepare image migration",
		Long: `Prepare a live migration.

Import-only mode:
  goceph rbd migration prepare --import-only --source-spec '<json>' <dest-image-spec>`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				if importOnly {
					if sourceSpec == "" {
						return fmt.Errorf("--source-spec is required with --import-only")
					}
					return conn.RbdMigrationPrepareImport(
						context.Background(),
						rbdapi.ImageSpec(args[0]),
						sourceSpec,
					)
				}

				src := rbdapi.ImageSpec(args[0])
				var dst rbdapi.ImageSpec
				if len(args) == 2 {
					dst = rbdapi.ImageSpec(args[1])
				}
				return conn.RbdMigrationPrepare(context.Background(), src, dst)
			})
		},
	}

	cmd.Flags().BoolVar(&importOnly, "import-only", false, "Prepare import-only migration")
	cmd.Flags().StringVar(&sourceSpec, "source-spec", "", "JSON-encoded source spec for import-only migration")

	return cmd
}

func newMigrationExecuteCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "execute <image-spec>",
		Short: "Execute image migration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				prog := rbdapi.NewProgress("Image migration", cmd.ErrOrStderr())
				return conn.RbdMigrationExecute(context.Background(), rbdapi.ImageSpec(args[0]), prog)
			})
		},
	}
}

func newMigrationCommitCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "commit <image-spec>",
		Short: "Commit image migration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdMigrationCommit(context.Background(), rbdapi.ImageSpec(args[0]))
			})
		},
	}
}

func newMigrationAbortCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "abort <image-spec>",
		Short: "Abort image migration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdMigrationAbort(context.Background(), rbdapi.ImageSpec(args[0]))
			})
		},
	}
}

func newMigrationStatusCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "status <image-spec>",
		Short: "Show image migration status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				status, err := conn.RbdMigrationStatus(context.Background(), rbdapi.ImageSpec(args[0]))
				if err != nil {
					return err
				}
				return app.WriteJSON(cmd.OutOrStdout(), status, true)
			})
		},
	}
}
