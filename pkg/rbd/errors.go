package rbd

import (
	"errors"

	cephrbd "github.com/ceph/go-ceph/rbd"
)

var (
	errInvalidImageSpec = errors.New("invalid image spec")
	errInvalidSnapSpec  = errors.New("invalid snap spec")
)

// ErrPlatformNotSupported is returned when RBD operations are called on unsupported platforms
var ErrPlatformNotSupported = errors.New("RBD is not supported on this platform")

func isErrNotFound(err error) bool {
	return errors.Is(err, cephrbd.ErrNotFound) || errors.Is(err, cephrbd.ErrNotExist) || errors.Is(err, cephrbd.RbdErrorNotFound)
}
