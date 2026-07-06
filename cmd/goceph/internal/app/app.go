package app

import (
	"context"
	"fmt"

	"github.com/bougou/go-ceph/pkg/rados"
	"github.com/spf13/cobra"
)

// Options holds global CLI flags shared across goceph subcommands.
type Options struct {
	CephConf string
	Retries  int
}

// BindFlags registers persistent flags on the root command.
func (o *Options) BindFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&o.CephConf, "conf", "c", "", "Ceph config path")
	cmd.PersistentFlags().IntVarP(&o.Retries, "retries", "r", 0, "Retry count for operations")
}

// WithConn runs fn with a RADOS connection.
func (o *Options) WithConn(ctx context.Context, fn func(*rados.RadosConn) error) error {
	conn, err := rados.NewRadosConn(o.CephConf)
	if err != nil {
		return fmt.Errorf("failed to create rados connection: %w", err)
	}
	defer conn.Close()
	conn.WithRetries(o.Retries)
	return fn(conn)
}

// WithoutConn runs fn without opening a RADOS connection.
func (o *Options) WithoutConn(ctx context.Context, fn func() error) error {
	return fn()
}
