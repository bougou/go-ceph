package rbd

import (
	"github.com/bougou/go-ceph/cmd/goceph/ceph/rbd/task"
	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/spf13/cobra"
)

// NewCmd builds the goceph ceph rbd command group.
func NewCmd(opts *app.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rbd",
		Short: "RBD mgr-module commands",
		Long:  "RBD operations exposed through the Ceph manager (not the standalone rbd binary).",
	}

	cmd.AddCommand(task.NewCmd(opts))

	return cmd
}
