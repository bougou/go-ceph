package main

import (
	"context"
	"fmt"
	"text/tabwriter"
	"time"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newSnapListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ls <image-spec>",
		Aliases: []string{"list"},
		Short:   "List snapshots",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				snaps, err := conn.RbdSnapList(context.Background(), ceph.ImageSpec(args[0]))
				if err != nil {
					return err
				}

				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "SNAPID\tNAME\tSIZE\tPROTECTED\tTIMESTAMP")
				for _, s := range snaps {
					protected := ""
					if s.Protected {
						protected = "yes"
					}

					timestamp := "-"
					if !s.Timestamp.IsZero() {
						timestamp = s.Timestamp.Local().Format(time.ANSIC)
					}

					fmt.Fprintf(
						w,
						"%d\t%s\t%s\t%s\t%s\n",
						s.ID,
						s.Name,
						s.SizeHuman(),
						protected,
						timestamp,
					)
				}
				_ = w.Flush()
				return nil
			})
		},
	}
}
