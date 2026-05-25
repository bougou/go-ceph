package rados

import (
	"context"

	"github.com/bougou/go-ceph/pkg/krbd"
	"github.com/bougou/go-ceph/pkg/rbd"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdChildren(ctx context.Context, imageSpec rbd.ImageSpec) (children []rbd.ImageSpec, err error) {
	err = rc.Do(ctx, func() error {
		_children, err := rbd.RbdChildren(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		children = _children
		return nil
	})
	return
}

func (rc *RadosConn) RbdSnapChildren(ctx context.Context, snapSpec rbd.SnapSpec) (children []rbd.ImageSpec, err error) {
	err = rc.Do(ctx, func() error {
		_children, err := rbd.RbdSnapChildren(ctx, rc.conn, snapSpec)
		if err != nil {
			return err
		}
		children = _children
		return nil
	})
	return
}

func (rc *RadosConn) RbdClone(ctx context.Context, srcSnapSpec rbd.SnapSpec, dstImageSpec rbd.ImageSpec, optFns ...rbd.RbdImageOptionFn) error {
	return rc.Do(ctx, func() error {
		return rbd.RbdClone(ctx, rc.conn, srcSnapSpec, dstImageSpec, optFns...)
	})
}

func (rc *RadosConn) RbdCopy(ctx context.Context, srcImageSpec rbd.ImageSpec, dstImageSpec rbd.ImageSpec, optFns ...rbd.RbdImageOptionFn) error {
	return rc.Do(ctx, func() error {
		return rbd.RbdCopy(ctx, rc.conn, srcImageSpec, dstImageSpec, optFns...)
	})
}

func (rc *RadosConn) RbdCreate(ctx context.Context, imageSpec rbd.ImageSpec, sizeBytes int64, optFns ...rbd.RbdImageOptionFn) error {
	return rc.Do(ctx, func() error {
		return rbd.RbdCreate(ctx, rc.conn, imageSpec, sizeBytes, optFns...)
	})
}

func (rc *RadosConn) RbdDeviceList(ctx context.Context) (devices []krbd.Device, err error) {
	err = rc.Do(ctx, func() error {
		_devices, err := rbd.RbdDeviceList(ctx, rc.conn)
		if err != nil {
			return err
		}
		devices = _devices
		return nil
	})
	return
}

func (rc *RadosConn) RbdDeviceMap(ctx context.Context, imageOrSnapSpec string, options *krbd.Options) error {
	return rc.Do(ctx, func() error {
		return rbd.RbdDeviceMap(ctx, rc.conn, imageOrSnapSpec, options)
	})
}

func (rc *RadosConn) RbdDeviceUnmap(ctx context.Context, imageOrSnapSpec string, options *krbd.Options) error {
	return rc.Do(ctx, func() error {
		return rbd.RbdDeviceUnmap(ctx, rc.conn, imageOrSnapSpec, options)
	})
}

func (rc *RadosConn) RbdDeviceUnmapByID(ctx context.Context, devID int, options *krbd.Options) error {
	return rc.Do(ctx, func() error {
		return rbd.RbdDeviceUnmapByID(ctx, rc.conn, devID, options)
	})
}

func (rc *RadosConn) RbdExist(ctx context.Context, imageSpec rbd.ImageSpec) (exist bool, err error) {
	err = rc.Do(ctx, func() error {
		_exist, err := rbd.RbdExist(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		exist = _exist
		return nil
	})
	return
}

func (rc *RadosConn) RbdFlatten(ctx context.Context, imageSpec rbd.ImageSpec) error {
	return rc.Do(ctx, func() error {
		return rbd.RbdFlatten(ctx, rc.conn, imageSpec)
	})
}

// RbdInfo retrieves detailed information about an RBD image.
// If the image does not exist, it returns nil, nil.
func (rc *RadosConn) RbdInfo(ctx context.Context, imageSpec rbd.ImageSpec) (info *rbd.ImageInfo, err error) {
	err = rc.Do(ctx, func() error {
		_info, err := rbd.RbdInfo(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		info = _info
		return nil
	})
	return
}

// RbdOpenImage opens an RBD image and returns it.
// You should close the image after using it.
func (rc *RadosConn) RbdOpenImage(ctx context.Context, imageSpec rbd.ImageSpec) (image *cephrbd.Image, err error) {
	err = rc.Do(ctx, func() error {
		_image, err := rbd.RbdOpenImage(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		image = _image
		return nil
	})
	return
}

func (rc *RadosConn) RbdRemove(ctx context.Context, imageSpec rbd.ImageSpec) error {
	return rc.Do(ctx, func() error {
		return rbd.RbdRemove(ctx, rc.conn, imageSpec)
	})
}

func (rc *RadosConn) RbdRename(ctx context.Context, srcImageSpec rbd.ImageSpec, dstImageSpec rbd.ImageSpec) error {
	return rc.Do(ctx, func() error {
		return rbd.RbdRename(ctx, rc.conn, srcImageSpec, dstImageSpec)
	})
}

func (rc *RadosConn) RbdResize(ctx context.Context, imageSpec rbd.ImageSpec, sizeBytes uint64) error {
	return rc.Do(ctx, func() error {
		return rbd.RbdResize(ctx, rc.conn, imageSpec, sizeBytes)
	})
}

func (rc *RadosConn) RbdSnapExist(ctx context.Context, snapSpec rbd.SnapSpec) (exist bool, err error) {
	err = rc.Do(ctx, func() error {
		_exist, err := rbd.RbdSnapExist(ctx, rc.conn, snapSpec)
		if err != nil {
			return err
		}
		exist = _exist
		return nil
	})
	return
}

func (rc *RadosConn) RbdSnapCreate(ctx context.Context, snapSpec rbd.SnapSpec) error {
	return rc.Do(ctx, func() error {
		return rbd.RbdSnapCreate(ctx, rc.conn, snapSpec)
	})
}

func (rc *RadosConn) RbdSnapRemove(ctx context.Context, snapSpec rbd.SnapSpec) error {
	return rc.Do(ctx, func() error {
		return rbd.RbdSnapRemove(ctx, rc.conn, snapSpec)
	})
}

func (rc *RadosConn) RbdSnapList(ctx context.Context, imageSpec rbd.ImageSpec) (snaps []rbd.SnapInfo, err error) {
	err = rc.Do(ctx, func() error {
		_snaps, err := rbd.RbdSnapList(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		snaps = _snaps
		return nil
	})
	return
}

func (rc *RadosConn) RbdStatus(ctx context.Context, imageOrSnapSpec string) (watchers []cephrbd.ImageWatcher, err error) {
	err = rc.Do(ctx, func() error {
		_watchers, err := rbd.RbdStatus(ctx, rc.conn, imageOrSnapSpec)
		if err != nil {
			return err
		}
		watchers = _watchers
		return nil
	})
	return
}
