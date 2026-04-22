package krbd

import (
	"fmt"
	"strings"
)

// Image is a Ceph RBD image.
type Image struct {
	DevID     int      // Unmap only
	Monitors  []string // Monitor endpoints for krbd map requests; host-only or v1-style host:port is generally the most compatible input.
	Namespace string
	Pool      string
	Image     string
	Snapshot  string
	Options   *Options
}

// MarshalText marshals Image attributes into the string format expected by the krbd add interface, e.g.
// "${mons} name=${user},secret=${key} ${pool} ${image} ${snap}".
func (i Image) MarshalText() (text []byte, err error) {
	if i.Snapshot == "" {
		i.Snapshot = "-"
	}

	options := "name=admin"
	if i.Options != nil {
		if s := i.Options.String(); s != "" {
			options = s
		}
	}

	text = []byte(fmt.Sprintf("%s %s %s %s %s", strings.Join(i.Monitors, ","), options, i.Pool, i.Image, i.Snapshot))
	return
}

func (i Image) String() string {
	b, err := i.MarshalText()
	if err != nil {
		return ""
	}
	return string(b)
}
