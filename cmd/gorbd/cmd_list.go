package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/bougou/go-ceph/pkg/rados"
	"github.com/bougou/go-ceph/pkg/rbd"
	"github.com/spf13/cobra"
)

const (
	listFormatPlain = "plain"
	listFormatJSON  = "json"
	listFormatXML   = "xml"
)

func newListCmd() *cobra.Command {
	var (
		long       bool
		poolFlag   string
		nsFlag     string
		formatFlag string
		pretty     bool
	)

	cmd := &cobra.Command{
		Use:     "list [<pool-spec>]",
		Aliases: []string{"ls"},
		Short:   "List rbd images",
		Long: `List rbd images.

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

			poolSpec, err := resolvePoolSpec(args, poolFlag, nsFlag)
			if err != nil {
				return err
			}

			return withConn(context.Background(), func(conn *rados.RadosConn) error {
				out := cmd.OutOrStdout()
				if long {
					entries, err := conn.RbdListLong(context.Background(), poolSpec)
					if err != nil {
						return err
					}
					return renderLong(out, entries, format, pretty)
				}
				images, err := conn.RbdList(context.Background(), poolSpec)
				if err != nil {
					return err
				}
				return renderShort(out, images, format, pretty)
			})
		},
	}

	cmd.Flags().BoolVarP(&long, "long", "l", false, "long listing format")
	cmd.Flags().StringVarP(&poolFlag, "pool", "p", "", "pool name")
	cmd.Flags().StringVar(&nsFlag, "namespace", "", "namespace name")
	cmd.Flags().StringVar(&formatFlag, "format", listFormatPlain, "output format (plain, json, or xml)")
	cmd.Flags().BoolVar(&pretty, "pretty-format", false, "pretty formatting (json and xml)")

	return cmd
}

// resolvePoolSpec merges the positional <pool-spec> with the --pool / --namespace flags.
// Positional wins if it carries a pool name; --namespace can still extend a positional
// that omits its namespace.
func resolvePoolSpec(args []string, poolFlag, nsFlag string) (rbd.PoolSpec, error) {
	var posPool, posNs string
	if len(args) == 1 {
		p, n, err := rbd.Pool(args[0])
		if err != nil {
			return "", err
		}
		posPool, posNs = p, n
	}

	pool := posPool
	if pool == "" {
		pool = poolFlag
	}
	namespace := posNs
	if namespace == "" {
		namespace = nsFlag
	}

	if pool == "" {
		return "", fmt.Errorf("pool name is required (positional <pool-spec> or --pool)")
	}

	spec := rbd.NewPoolSpec(pool, namespace)
	if !spec.Valid() {
		return "", fmt.Errorf("invalid pool spec: %s", spec)
	}
	return spec, nil
}

func renderShort(w io.Writer, images []string, format string, pretty bool) error {
	switch format {
	case listFormatJSON:
		return writeJSON(w, images, pretty)
	case listFormatXML:
		return writeImagesXML(w, images, pretty)
	default:
		for _, name := range images {
			fmt.Fprintln(w, name)
		}
		return nil
	}
}

func renderLong(w io.Writer, entries []rbd.ImageListEntry, format string, pretty bool) error {
	switch format {
	case listFormatJSON:
		return writeJSON(w, entries, pretty)
	case listFormatXML:
		return writeEntriesXML(w, entries, pretty)
	default:
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tSIZE\tPARENT\tFMT\tPROT\tLOCK")
		for _, e := range entries {
			name := e.Image
			if e.Snapshot != "" {
				name = e.Image + "@" + e.Snapshot
			}
			parent := e.Parent
			if parent == "" {
				parent = ""
			}
			protected := e.Protected
			if e.Snapshot == "" {
				protected = ""
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t\n", name, humanSize(e.Size), parent, e.Format, protected)
		}
		return tw.Flush()
	}
}

func writeJSON(w io.Writer, v any, pretty bool) error {
	enc := json.NewEncoder(w)
	if pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(v)
}

// xmlImage and xmlEntries are small wrappers so encoding/xml emits a stable root element.
type xmlImage struct {
	XMLName xml.Name `xml:"image"`
	Name    string   `xml:",chardata"`
}

type xmlImageList struct {
	XMLName xml.Name   `xml:"images"`
	Images  []xmlImage `xml:"image"`
}

type xmlEntryList struct {
	XMLName xml.Name             `xml:"images"`
	Entries []rbd.ImageListEntry `xml:"image"`
}

func writeImagesXML(w io.Writer, images []string, pretty bool) error {
	list := xmlImageList{Images: make([]xmlImage, 0, len(images))}
	for _, name := range images {
		list.Images = append(list.Images, xmlImage{Name: name})
	}
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

func writeEntriesXML(w io.Writer, entries []rbd.ImageListEntry, pretty bool) error {
	list := xmlEntryList{Entries: entries}
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

func humanSize(size uint64) string {
	const (
		KiB = 1024
		MiB = KiB * 1024
		GiB = MiB * 1024
		TiB = GiB * 1024
	)
	switch {
	case size >= TiB:
		return fmt.Sprintf("%d TiB", size/TiB)
	case size >= GiB:
		return fmt.Sprintf("%d GiB", size/GiB)
	case size >= MiB:
		return fmt.Sprintf("%d MiB", size/MiB)
	case size >= KiB:
		return fmt.Sprintf("%d KiB", size/KiB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
