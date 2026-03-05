package ceph

import (
	"errors"

	"github.com/ceph/go-ceph/rbd"
)

var (
	errInvalidImageSpec = errors.New("invalid image spec")
	errInvalidSnapSpec  = errors.New("invalid snap spec")
)

func isErrNotFound(err error) bool {
	return errors.Is(err, rbd.ErrNotFound) || errors.Is(err, rbd.ErrNotExist) || errors.Is(err, rbd.RbdErrorNotFound)
}
