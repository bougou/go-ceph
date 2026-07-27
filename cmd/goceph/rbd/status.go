package rbd

import (
	"context"
	"fmt"
	"time"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newStatusCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "status <image-or-snap-spec>",
		Short: "Show image or snapshot status",
		Long: `Show watchers, migration status, and persistent cache state.

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

				if c := status.PersistentCache; c != nil {
					fmt.Fprintln(out, "Persistent cache state:")
					fmt.Fprintf(out, "\thost: %s\n", c.Host)
					fmt.Fprintf(out, "\tpath: %s\n", c.Path)
					fmt.Fprintf(out, "\tsize: %s\n", rbdapi.ByteSizeHuman(c.Size))
					fmt.Fprintf(out, "\tmode: %s\n", c.Mode)
					fmt.Fprintf(out, "\tstats_timestamp: %s\n", c.StatsTimestamp.Format(time.ANSIC))
					fmt.Fprintf(
						out,
						"\tpresent: %t\tempty: %t\tclean: %t\n",
						c.Present, c.Empty, c.Clean,
					)
					fmt.Fprintf(out, "\tallocated: %s\n", rbdapi.ByteSizeHuman(c.AllocatedBytes))
					fmt.Fprintf(out, "\tcached: %s\n", rbdapi.ByteSizeHuman(c.CachedBytes))
					fmt.Fprintf(out, "\tdirty: %s\n", rbdapi.ByteSizeHuman(c.DirtyBytes))
					fmt.Fprintf(out, "\tfree: %s\n", rbdapi.ByteSizeHuman(c.FreeBytes))
					fmt.Fprintf(out, "\thits_full: %d / %d%%\n", c.HitsFull, c.HitsFullPercent())
					fmt.Fprintf(out, "\thits_partial: %d / %d%%\n", c.HitsPartial, c.HitsPartialPercent())
					fmt.Fprintf(out, "\tmisses: %d\n", c.Misses)
					fmt.Fprintf(out, "\thit_bytes: %s / %d%%\n", rbdapi.ByteSizeHuman(c.HitBytes), c.HitBytesPercent())
					fmt.Fprintf(out, "\tmiss_bytes: %s\n", rbdapi.ByteSizeHuman(c.MissBytes))
				}
				return nil
			})
		},
	}
}
