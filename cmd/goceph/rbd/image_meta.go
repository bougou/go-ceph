package rbd

import (
	"context"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newImageMetaCmd(opts *app.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image-meta",
		Short: "Image metadata operations",
		Long:  "Get, list, set, or remove RBD image metadata key/value pairs.",
	}
	cmd.AddCommand(
		newImageMetaGetCmd(opts),
		newImageMetaListCmd(opts),
		newImageMetaSetCmd(opts),
		newImageMetaRemoveCmd(opts),
	)
	return cmd
}

func newImageMetaGetCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <image-spec> <key>",
		Short: "Get image metadata value for a key",
		Long: `Image metadata get the value associated with the key.

Positional arguments:
  <image-spec>  image specification
                ([<pool-name>/[<namespace>/]]<image-name>)
  <key>         image meta key`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				value, err := conn.RbdImageMetaGet(context.Background(), rbdapi.ImageSpec(args[0]), args[1])
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), value)
				return nil
			})
		},
	}
}

func newImageMetaListCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:     "list <image-spec>",
		Aliases: []string{"ls"},
		Short:   "List image metadata keys with values",
		Long: `Image metadata list keys with values.

Positional arguments:
  <image-spec>  image specification
                ([<pool-name>/[<namespace>/]]<image-name>)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				meta, err := conn.RbdImageMetaList(context.Background(), rbdapi.ImageSpec(args[0]))
				if err != nil {
					return err
				}
				return renderImageMetaList(cmd.OutOrStdout(), meta)
			})
		},
	}
}

func newImageMetaSetCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "set <image-spec> <key> <value>",
		Short: "Set image metadata key with value",
		Long: `Image metadata set key with value.

Positional arguments:
  <image-spec>  image specification
                ([<pool-name>/[<namespace>/]]<image-name>)
  <key>         image meta key
  <value>       image meta value`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdImageMetaSet(context.Background(), rbdapi.ImageSpec(args[0]), args[1], args[2])
			})
		},
	}
}

func newImageMetaRemoveCmd(opts *app.Options) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <image-spec> <key>",
		Aliases: []string{"rm"},
		Short:   "Remove image metadata key and value",
		Long: `Image metadata remove the key and value associated.

Positional arguments:
  <image-spec>  image specification
                ([<pool-name>/[<namespace>/]]<image-name>)
  <key>         image meta key`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				return conn.RbdImageMetaRemove(context.Background(), rbdapi.ImageSpec(args[0]), args[1])
			})
		},
	}
}

func renderImageMetaList(w io.Writer, meta map[string]string) error {
	if len(meta) == 0 {
		return nil
	}

	keys := make([]string, 0, len(meta))
	for key := range meta {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Key\tValue")
	for _, key := range keys {
		fmt.Fprintf(tw, "%s\t%s\n", key, meta[key])
	}
	return tw.Flush()
}
