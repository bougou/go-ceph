package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	ceph "github.com/bougou/go-ceph"
	"github.com/bougou/go-ceph/krbd"
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
		newDeviceMapCmd(),
		newDeviceUnmapCmd(),
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

func newDeviceMapCmd() *cobra.Command {
	var (
		deviceType string
		cookie     string
		showCookie bool
		snapID     int64
		readOnly   bool
		exclusive  bool
		optionsStr string
	)

	cmd := &cobra.Command{
		Use:   "map <image-spec|snap-spec>",
		Short: "Map RBD image or snapshot to a block device",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceType == "" {
				deviceType = "krbd"
			}
			if deviceType != "krbd" {
				return fmt.Errorf("unsupported device type %q, only krbd is supported", deviceType)
			}
			if cookie != "" || showCookie || snapID != 0 {
				return errors.New("--cookie, --show-cookie and --snap-id are not supported for krbd")
			}

			options := &krbd.Options{}
			if err := options.UnmarshalText([]byte(optionsStr)); err != nil {
				return err
			}
			if readOnly {
				options.ReadOnly = true
			}
			if exclusive {
				options.Exclusive = true
			}

			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				return conn.RbdDeviceMap(context.Background(), args[0], options)
			})
		},
	}

	cmd.Flags().StringVarP(&deviceType, "device-type", "t", "krbd", "Device type (only krbd is supported)")
	cmd.Flags().StringVar(&cookie, "cookie", "", "Device cookie (not supported for krbd)")
	cmd.Flags().BoolVar(&showCookie, "show-cookie", false, "Show generated cookie (not supported for krbd)")
	cmd.Flags().Int64Var(&snapID, "snap-id", 0, "Snapshot id (not supported for krbd)")
	cmd.Flags().BoolVar(&readOnly, "read-only", false, "Map as read-only")
	cmd.Flags().BoolVar(&exclusive, "exclusive", false, "Map as exclusive")
	cmd.Flags().StringVarP(&optionsStr, "options", "o", "", "Comma separated device options (opt1,opt2=val,...)")

	return cmd
}

func newDeviceUnmapCmd() *cobra.Command {
	var (
		deviceType string
		optionsStr string
		snapID     int64
	)

	cmd := &cobra.Command{
		Use:   "unmap <image-spec|snap-spec|device-path>",
		Short: "Unmap by image spec, snapshot spec, or device path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if deviceType == "" {
				deviceType = "krbd"
			}
			if deviceType != "krbd" {
				return fmt.Errorf("unsupported device type %q, only krbd is supported", deviceType)
			}
			if snapID != 0 {
				return errors.New("--snap-id is not supported for krbd unmap")
			}

			unmapOpts := &krbd.Options{}
			if err := unmapOpts.UnmarshalText([]byte(optionsStr)); err != nil {
				return err
			}

			return withConn(context.Background(), func(conn *ceph.RadosConn) error {
				input := strings.TrimSpace(args[0])
				if devID, ok := parseDeviceIDFromPath(input); ok {
					return conn.RbdDeviceUnmapByID(context.Background(), devID, unmapOpts)
				}
				return conn.RbdDeviceUnmap(context.Background(), input, unmapOpts)
			})
		},
	}

	cmd.Flags().StringVarP(&deviceType, "device-type", "t", "krbd", "Device type (only krbd is supported)")
	cmd.Flags().StringVarP(&optionsStr, "options", "o", "", "Comma separated device options (opt1,opt2=val,...)")
	cmd.Flags().Int64Var(&snapID, "snap-id", 0, "Snapshot id (not supported for krbd)")

	return cmd
}

func parseDeviceIDFromPath(path string) (id int, ok bool) {
	s := strings.TrimSpace(path)
	if s == "" {
		return
	}
	id, err := strconv.Atoi(s)
	if err == nil {
		ok = true
		return
	}

	base := filepath.Base(s)
	if strings.HasPrefix(base, "rbd") {
		idStr := strings.TrimPrefix(base, "rbd")
		id, err = strconv.Atoi(idStr)
		if err == nil {
			ok = true
			return
		}
	}
	return
}
