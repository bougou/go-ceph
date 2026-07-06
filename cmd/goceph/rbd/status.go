package rbd

import (
	"context"
	"fmt"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newStatusCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "status <image-or-snap-spec>",
		Short: "Show image or snapshot status",
		Long: `Show watchers and migration status for an image or snapshot.

Positional arguments:
  <image-or-snap-spec>  image or snapshot specification
                        ([<pool-name>/[<namespace>/]]<image-name>[@<snap-name>])`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				status, err := conn.RbdStatus(context.Background(), args[0])
				if err != nil {
					return err
				}

				out := cmd.OutOrStdout()
				if len(status.Watchers) == 0 {
					fmt.Fprintln(out, "Watchers: none")
				} else {
					fmt.Fprintln(out, "Watchers:")
					for _, watcher := range status.Watchers {
						fmt.Fprintf(
							out,
							"\twatcher=%s client.%d cookie=%d\n",
							watcher.Addr,
							watcher.Id,
							watcher.Cookie,
						)
					}
				}

				if status.Migration != nil {
					fmt.Fprintln(out, "Migration:")
					fmt.Fprintf(out, "\tsource: %s\n", status.Migration.Source)
					fmt.Fprintf(out, "\tdestination: %s\n", status.Migration.Destination)
					fmt.Fprintf(out, "\tstate: %s\n", rbdapi.MigrationStateLine(status.Migration))
				}
				return nil
			})
		},
	}
}
