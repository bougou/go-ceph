package rbd

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/bougou/go-ceph/cmd/goceph/internal/app"
	"github.com/bougou/go-ceph/pkg/rados"
	rbdapi "github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

func newTrashCmd(opts *app.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trash",
		Short: "Manage deleted images",
	}
	cmd.AddCommand(newTrashListCmd(opts))
	return cmd
}

func newTrashListCmd(opts *app.Options) *cobra.Command {
	var (
		long       bool
		all        bool
		poolFlag   string
		nsFlag     string
		formatFlag string
		pretty     bool
	)

	cmd := &cobra.Command{
		Use:     "list [<pool-spec>]",
		Aliases: []string{"ls"},
		Short:   "List trash images",
		Long: `List trash images.

Positional arguments:
  <pool-spec>  pool specification (example: <pool-name>[/<namespace>])`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format := strings.ToLower(strings.TrimSpace(formatFlag))
			switch format {
			case "", listFormatPlain:
				format = listFormatPlain
			case listFormatJSON, listFormatXML:
			default:
				return fmt.Errorf("invalid --format %q (allowed: plain, json, xml)", formatFlag)
			}

			poolSpec, err := resolvePoolSpecOptional(args, poolFlag, nsFlag)
			if err != nil {
				return err
			}

			return opts.WithConn(context.Background(), func(conn *rados.RadosConn) error {
				_ = all
				entries, err := conn.RbdTrashList(context.Background(), poolSpec)
				if err != nil {
					return err
				}
				return renderTrashList(cmd.OutOrStdout(), entries, format, pretty, long)
			})
		},
	}

	cmd.Flags().BoolVarP(&long, "long", "l", false, "long listing format")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "list images from all sources")
	cmd.Flags().StringVarP(&poolFlag, "pool", "p", "", "pool name")
	cmd.Flags().StringVar(&nsFlag, "namespace", "", "namespace name")
	cmd.Flags().StringVar(&formatFlag, "format", listFormatPlain, "output format (plain, json, or xml)")
	cmd.Flags().BoolVar(&pretty, "pretty-format", false, "pretty formatting (json and xml)")

	return cmd
}

func resolvePoolSpecOptional(args []string, poolFlag, nsFlag string) (rbdapi.PoolSpec, error) {
	var posPool, posNs string
	if len(args) == 1 {
		p, n, err := rbdapi.Pool(args[0])
		if err != nil {
			return "", err
		}
		posPool, posNs = p, n
	}

	pool := posPool
	if pool == "" {
		pool = poolFlag
	}
	if pool == "" {
		pool = rbdapi.DefaultPoolName
	}

	namespace := posNs
	if namespace == "" {
		namespace = nsFlag
	}

	spec := rbdapi.NewPoolSpec(pool, namespace)
	if !spec.Valid() {
		return "", fmt.Errorf("invalid pool spec: %s", spec)
	}
	return spec, nil
}

func renderTrashList(w io.Writer, entries []rbdapi.Trash, format string, pretty, long bool) error {
	switch format {
	case listFormatJSON:
		payload := struct {
			Trash []rbdapi.Trash `json:"trash"`
		}{Trash: entries}
		return app.WriteJSON(w, payload, pretty)
	case listFormatXML:
		return writeTrashListXML(w, entries, pretty)
	default:
		if long {
			return renderTrashListLongPlain(w, entries)
		}
		for _, entry := range entries {
			fmt.Fprintf(w, "%s %s\n", entry.ID, entry.Name)
		}
		return nil
	}
}

func renderTrashListLongPlain(w io.Writer, entries []rbdapi.Trash) error {
	if len(entries) == 0 {
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tDELETED_AT\tSTATUS\tPARENT")
	for _, entry := range entries {
		parent := ""
		if entry.Parent != nil {
			parent = entry.Parent.Image.PoolName + "/"
			if entry.Parent.Image.PoolNamespace != "" {
				parent += entry.Parent.Image.PoolNamespace + "/"
			}
			parent += entry.Parent.Image.ImageName + "@" + entry.Parent.Snap.SnapName
		}
		deletedAt := ""
		if !entry.DeletedAt.IsZero() {
			deletedAt = entry.DeletedAt.Local().Format(time.ANSIC)
		}
		fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\t%s\n",
			entry.ID,
			entry.Name,
			deletedAt,
			entry.Status,
			parent,
		)
	}
	return tw.Flush()
}

type xmlTrashList struct {
	XMLName xml.Name       `xml:"trash"`
	Entries []rbdapi.Trash `xml:"image"`
}

func writeTrashListXML(w io.Writer, entries []rbdapi.Trash, pretty bool) error {
	list := xmlTrashList{Entries: entries}
	enc := xml.NewEncoder(w)
	if pretty {
		enc.Indent("", "  ")
	}
	if err := enc.Encode(list); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w)
	return err
}
