package root

import (
	"github.com/bougou/go-ceph/cmd/goceph/ceph"
	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/cmd/goceph/rados"
	"github.com/bougou/go-ceph/cmd/goceph/rbd"
	"github.com/spf13/cobra"
)

// NewCmd builds the goceph root command.
func NewCmd() *cobra.Command {
	opts := &app.Options{}

	cmd := &cobra.Command{
		Use:   "goceph",
		Short: "Ceph helper CLI based on go-ceph",
		Long: `goceph groups commands by the native Ceph CLI they mirror:

  goceph rbd    — RBD image operations (native rbd)
  goceph ceph   — Ceph cluster / mgr operations (native ceph)
  goceph rados  — RADOS operations (native rados)`,
	}

	opts.BindFlags(cmd)
	cmd.AddCommand(
		rbd.NewCmd(opts),
		ceph.NewCmd(opts),
		rados.NewCmd(opts),
	)

	return cmd
}

// Execute runs the goceph CLI.
func Execute() error {
	return NewCmd().Execute()
}
