package rbd

import (
	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/spf13/cobra"
)

func newSnapCmd(opts *app.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snap",
		Short: "Snapshot operations",
	}
	cmd.AddCommand(
		newSnapListCmd(opts),
		newSnapCreateCmd(opts),
		newSnapRemoveCmd(opts),
	)
	return cmd
}
