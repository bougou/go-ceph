package krbd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Reference: https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-bus-rbd

const SysBusRbdPath = "/sys/bus/rbd"
const SysModuleRbdParametersPath = "/sys/module/rbd/parameters"

type Parameters struct {
	SingleMajor bool `krbd:"single_major"`
}

type parameterTag struct {
	name     string
	optional bool
}

func parseSysfsBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "y", "yes", "true", "on":
		return true, nil
	case "0", "n", "no", "false", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool value %q", value)
	}
}

func parseParameterTag(field reflect.StructField) (p parameterTag) {
	tag := field.Tag.Get("krbd")
	if tag == "" {
		return
	}
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return
	}
	p.name = parts[0]
	for i := 1; i < len(parts); i++ {
		switch parts[i] {
		case "optional":
			p.optional = true
		}
	}
	return
}

func (p *Parameters) decode(path string) error {
	t := reflect.TypeOf(*p)
	v := reflect.ValueOf(p).Elem()

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("krbd")
		if tag == "" {
			continue
		}
		parameterTag := parseParameterTag(t.Field(i))
		if parameterTag.name == "" {
			continue
		}

		filePath := path + "/" + parameterTag.name
		r, err := os.Open(filePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) && parameterTag.optional {
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
			switch parameterTag.name {
			case "single_major":
				b, err := parseSysfsBool(string(value))
				if err != nil {
					return fmt.Errorf("failed to parse bool parameter (%s=%s): %w", parameterTag.name, string(value), err)
				}
				v.Field(i).SetBool(b)
			}
		}
	}

	return nil
}

func MapWriter() (io.WriteCloser, error) {
	parameters := Parameters{}
	err := parameters.decode(SysModuleRbdParametersPath)
	if err != nil {
		return nil, fmt.Errorf("failed to decode parameters: %w", err)
	}

	if parameters.SingleMajor {
		path := filepath.Join(SysBusRbdPath, "add_single_major")
		w, err := os.OpenFile(path, os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open path (%s): %w", path, err)
		}
		return w, nil
	}

	path := filepath.Join(SysBusRbdPath, "add")
	w, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open path (%s): %w", path, err)
	}
	return w, nil
}

func UnmapWriter() (io.WriteCloser, error) {
	parameters := Parameters{}
	err := parameters.decode(SysModuleRbdParametersPath)
	if err != nil {
		return nil, fmt.Errorf("failed to decode parameters: %w", err)
	}

	if parameters.SingleMajor {
		path := filepath.Join(SysBusRbdPath, "remove_single_major")
		w, err := os.OpenFile(path, os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open path (%s): %w", path, err)
		}
		return w, nil
	}

	path := filepath.Join(SysBusRbdPath, "remove")
	w, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open path (%s): %w", path, err)
	}
	return w, nil
}
