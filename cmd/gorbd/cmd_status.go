package main

import (
	"context"
	"fmt"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <image-or-snap-spec>",
		Short: "Show watchers of an image or snapshot",
		Long: `Show watchers of an image or snapshot.

Positional arguments:
  <image-or-snap-spec>  image or snapshot specification
                        ([<pool-name>/[<namespace>/]]<image-name>[@<snap-name>])`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				watchers, err := conn.RbdStatus(context.Background(), args[0])
				if err != nil {
					return err
				}

				out := cmd.OutOrStdout()
				if len(watchers) == 0 {
					fmt.Fprintln(out, "Watchers: none")
					return nil
				}

				fmt.Fprintln(out, "Watchers:")
				for _, watcher := range watchers {
					fmt.Fprintf(
						out,
						"\twatcher=%s client.%d cookie=%d\n",
						watcher.Addr,
						watcher.Id,
						watcher.Cookie,
					)
				}
				return nil
			})
		},
	}
}
