package krbd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/ceph/go-ceph/rbd"
)

const SysBusRbdDevicesPath = "/sys/bus/rbd/devices"

type Device struct {
	ID        int64
	Pool      string `krbd:"pool"`
	Namespace string `krbd:"pool_ns,optional"`
	Image     string `krbd:"name"`
	Snapshot  string `krbd:"current_snap,optional"` // Note, the kernel may expose the snapshot name as "-" when there is no snapshot. But we decode it to empty string to represent no snapshot.
	Size      string `krbd:"size"`
	Features  uint64 `krbd:"features"`
}

type deviceTag struct {
	name     string
	optional bool
}

func parseDeviceTag(field reflect.StructField) (d deviceTag) {
	tag := field.Tag.Get("krbd")
	if tag == "" {
		return
	}
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return
	}
	d.name = parts[0]
	for i := 1; i < len(parts); i++ {
		switch parts[i] {
		case "optional":
			d.optional = true
		}
	}
	return
}

func parseFeaturesValue(value string) (features uint64, err error) {
	features, err = strconv.ParseUint(value, 0, 64)
	return
}

func (d *Device) decode(path string) error {
	t := reflect.TypeOf(*d)
	v := reflect.ValueOf(d).Elem()

	for i := 0; i < t.NumField(); i++ {
		deviceTag := parseDeviceTag(t.Field(i))
		if deviceTag.name == "" {
			continue
		}

		filePath := path + "/" + deviceTag.name
		r, err := os.Open(filePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) && deviceTag.optional {
				continue
			}
			return fmt.Errorf("failed to open file (%s): %w", filePath, err)
		}
		defer r.Close()

		value, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read file (%s): %w", filePath, err)
		}

		value = bytes.TrimSpace(value)

		if len(value) != 0 {
			switch deviceTag.name {
			case "features":
				features, err := parseFeaturesValue(string(value))
				if err != nil {
					return fmt.Errorf("failed to parse features value (%s): %w", string(value), err)
				}
				d.Features = features
				// v.Field(i).SetUint(features)
			case "current_snap":
				if string(value) == "-" {
					d.Snapshot = ""
				} else {
					d.Snapshot = string(value)
				}
			default:
				v.Field(i).SetString(string(value))
			}
		}
	}

	return nil
}

func (d *Device) FeatureNames() []string {
	featureSet := rbd.FeatureSet(d.Features)
	names := featureSet.Names() // Names() is backed by a map, so the order is non-deterministic.
	sort.Strings(names)         // Sort to make the output deterministic.
	return names
}

// Devices iterates over all RBD devices and returns a list of Device structs.
func Devices() (devices []Device, err error) {
	entries, err := os.ReadDir(SysBusRbdDevicesPath)
	if err != nil {
		err = fmt.Errorf("failed to read directory: %w", err)
		return
	}

	devices = make([]Device, len(entries))
	for _, entry := range entries {
		id, parseErr := strconv.ParseInt(entry.Name(), 10, 0)
		if parseErr != nil {
			err = fmt.Errorf("failed to parse device id from entry name (%s): %w", entry.Name(), parseErr)
			return
		}

		device := Device{
			ID: id,
		}
		err = device.decode(SysBusRbdDevicesPath + "/" + entry.Name())
		if err != nil {
			err = fmt.Errorf("failed to decode device (%s): %w", entry.Name(), err)
			return
		}
		devices[id] = device
	}

	sort.SliceStable(devices, func(i, j int) bool {
		return devices[i].ID < devices[j].ID
	})

	return
}

func (d *Device) DevPath() string {
	return fmt.Sprintf("%s%d", "/dev/rbd", d.ID)
}

func Find(namespace, pool, image, snapshot string) (device Device, err error) {
	devices, err := Devices()
	if err != nil {
		err = fmt.Errorf("failed to get devices: %w", err)
		return
	}

	for _, d := range devices {
		if d.Namespace == namespace && d.Pool == pool && d.Image == image && d.Snapshot == snapshot {
			device = d
			return
		}
	}

	err = fmt.Errorf("device not found")
	return
}
