package krbd

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

// Map the RBD image via the krbd interface. An open io.Writer is required
// typically to /sys/bus/rbd/add or /sys/bus/rbd/add_single_major
func (i *Image) Map(w io.Writer) error {
	if len(i.Monitors) == 0 {
		return errors.New("no monitors defined")
	}
	if i.Pool == "" {
		return errors.New("no pool defined")
	}
	if i.Image == "" {
		return errors.New("no image defined")
	}

	out := i.String()
	n, err := w.Write([]byte(out))
	if err != nil {
		return fmt.Errorf("failed to write map request: %w", err)
	}

	if n != len(out) {
		return fmt.Errorf("incomplete write, wrote %d, expected to write %d", n, len(out))
	}
	return nil
}

// Unmap a RBD device via the krbd interface. DevID must be defined.
// An open io.Writer is required typically to /sys/bus/rbd/remove
// or /sys/bus/rbd/remove_single_major.
func (i *Image) Unmap(w io.Writer) error {
	cmd := strconv.Itoa(i.DevID)
	if i.Options != nil && i.Options.Force {
		cmd = cmd + " force"
	}
	n, err := w.Write([]byte(cmd))
	if err != nil {
		return fmt.Errorf("failed to write unmap request: %w", err)
	}
	if n != len(cmd) {
		return fmt.Errorf("incomplete write, wrote %d, expected to write %d", n, len(cmd))
	}
	return nil
}
