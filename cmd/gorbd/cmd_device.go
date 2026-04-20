package main

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	ceph "github.com/bougou/go-ceph"
	"github.com/spf13/cobra"
)

func newDeviceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "device",
		Short: "Device operations",
	}
	cmd.AddCommand(
		newDeviceListCmd(),
		newDeviceFindCmd(),
	)
	return cmd
}

func newDeviceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List mapped RBD devices",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withoutConn(context.Background(), func() error {
				devices, err := ceph.RbdDeviceList(context.Background(), nil)
				if err != nil {
					return err
				}

				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "id\tpool\tnamespace\timage\tsnap\tdevice\tfeatures")
				for _, d := range devices {
					snapshot := d.Snapshot
					if snapshot == "" {
						snapshot = "-"
					}
					featureNames := strings.Join(d.FeatureNames(), ",")
					fmt.Fprintf(
						w,
						"%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
						d.ID,
						d.Pool,
						d.Namespace,
						d.Image,
						snapshot,
						d.DevPath(),
						featureNames,
					)
				}
				_ = w.Flush()
				return nil
			})
		},
	}
}

func newDeviceFindCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "find <image-spec|snap-spec>",
		Short: "Find mapped RBD device by image or snapshot spec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withoutConn(context.Background(), func() error {
				device, err := ceph.RbdDeviceFind(context.Background(), nil, args[0])
				if err != nil {
					return err
				}

				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "id\tpool\tnamespace\timage\tsnap\tdevice\tfeatures")
				snapshot := device.Snapshot
				if snapshot == "" {
					snapshot = "-"
				}
				featureNames := strings.Join(device.FeatureNames(), ",")
				fmt.Fprintf(
					w,
					"%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
					device.ID,
					device.Pool,
					device.Namespace,
					device.Image,
					snapshot,
					device.DevPath(),
					featureNames,
				)
				_ = w.Flush()
				return nil
			})
		},
	}
}
