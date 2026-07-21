package rados

import (
	"context"

	"github.com/bougou/go-ceph/pkg/krbd"
	"github.com/bougou/go-ceph/pkg/rbd"
	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdParents(ctx context.Context, imageSpec rbd.ImageSpec) (parents []*cephrbd.ParentInfo, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		_parents, err := rbd.RbdParents(ctx, conn, imageSpec)
		if err != nil {
			return err
		}
		parents = _parents
		return nil
	})
	return
}

func (rc *RadosConn) RbdSnapParents(ctx context.Context, snapSpec rbd.SnapSpec) (parents []*cephrbd.ParentInfo, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		_parents, err := rbd.RbdSnapParents(ctx, conn, snapSpec)
		if err != nil {
			return err
		}
		parents = _parents
		return nil
	})
	return
}

func (rc *RadosConn) RbdChildren(ctx context.Context, imageSpec rbd.ImageSpec) (children []rbd.ImageSpec, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		_children, err := rbd.RbdChildren(ctx, conn, imageSpec)
		if err != nil {
			return err
		}
		children = _children
		return nil
	})
	return
}

func (rc *RadosConn) RbdSnapChildren(ctx context.Context, snapSpec rbd.SnapSpec) (children []rbd.ImageSpec, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		_children, err := rbd.RbdSnapChildren(ctx, conn, snapSpec)
		if err != nil {
			return err
		}
		children = _children
		return nil
	})
	return
}

func (rc *RadosConn) RbdClone(ctx context.Context, srcSnapSpec rbd.SnapSpec, dstImageSpec rbd.ImageSpec, opts ...rbd.CloneOption) (task *rbd.FlattenTask, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		task, err = rbd.RbdClone(ctx, conn, srcSnapSpec, dstImageSpec, opts...)
		return err
	})
	return
}

func (rc *RadosConn) RbdCopy(ctx context.Context, srcImageSpec rbd.ImageSpec, dstImageSpec rbd.ImageSpec, opts ...rbd.CopyOption) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdCopy(ctx, conn, srcImageSpec, dstImageSpec, opts...)
	})
}

func (rc *RadosConn) RbdCopySnap(ctx context.Context, srcSnapSpec rbd.SnapSpec, dstImageSpec rbd.ImageSpec, opts ...rbd.CopyOption) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdCopySnap(ctx, conn, srcSnapSpec, dstImageSpec, opts...)
	})
}

func (rc *RadosConn) RbdCopyUnsafe(ctx context.Context, srcImageSpec rbd.ImageSpec, dstImageSpec rbd.ImageSpec) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdCopyUnsafe(ctx, conn, srcImageSpec, dstImageSpec)
	})
}

func (rc *RadosConn) RbdCreate(ctx context.Context, imageSpec rbd.ImageSpec, sizeBytes int64, optFns ...rbd.RbdImageOption) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdCreate(ctx, conn, imageSpec, sizeBytes, optFns...)
	})
}

func (rc *RadosConn) RbdDeviceList(ctx context.Context) (devices []krbd.Device, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		_devices, err := rbd.RbdDeviceList(ctx, conn)
		if err != nil {
			return err
		}
		devices = _devices
		return nil
	})
	return
}

func (rc *RadosConn) RbdDeviceMap(ctx context.Context, imageOrSnapSpec string, options *krbd.Options) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdDeviceMap(ctx, conn, imageOrSnapSpec, options)
	})
}

func (rc *RadosConn) RbdDeviceUnmap(ctx context.Context, imageOrSnapSpec string, options *krbd.Options) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdDeviceUnmap(ctx, conn, imageOrSnapSpec, options)
	})
}

func (rc *RadosConn) RbdDeviceUnmapByID(ctx context.Context, devID int, options *krbd.Options) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdDeviceUnmapByID(ctx, conn, devID, options)
	})
}

func (rc *RadosConn) RbdExist(ctx context.Context, imageSpec rbd.ImageSpec) (exist bool, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		_exist, err := rbd.RbdExist(ctx, conn, imageSpec)
		if err != nil {
			return err
		}
		exist = _exist
		return nil
	})
	return
}

func (rc *RadosConn) RbdFlatten(ctx context.Context, imageSpec rbd.ImageSpec) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdFlatten(ctx, conn, imageSpec)
	})
}

// RbdInfo retrieves detailed information about an RBD image.
// If the image does not exist, it returns nil, nil.
func (rc *RadosConn) RbdInfo(ctx context.Context, imageSpec rbd.ImageSpec) (info *rbd.ImageInfo, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		_info, err := rbd.RbdInfo(ctx, conn, imageSpec)
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
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		_image, err := rbd.RbdOpenImage(ctx, conn, imageSpec)
		if err != nil {
			return err
		}
		image = _image
		return nil
	})
	return
}

func (rc *RadosConn) RbdRemove(ctx context.Context, imageSpec rbd.ImageSpec) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdRemove(ctx, conn, imageSpec)
	})
}

func (rc *RadosConn) RbdForceRemove(ctx context.Context, imageSpec rbd.ImageSpec) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdForceRemove(ctx, conn, imageSpec)
	})
}

func (rc *RadosConn) RbdRename(ctx context.Context, srcImageSpec rbd.ImageSpec, dstImageSpec rbd.ImageSpec) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdRename(ctx, conn, srcImageSpec, dstImageSpec)
	})
}

func (rc *RadosConn) RbdResize(ctx context.Context, imageSpec rbd.ImageSpec, sizeBytes uint64) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdResize(ctx, conn, imageSpec, sizeBytes)
	})
}

func (rc *RadosConn) RbdSnapExist(ctx context.Context, snapSpec rbd.SnapSpec) (exist bool, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		_exist, err := rbd.RbdSnapExist(ctx, conn, snapSpec)
		if err != nil {
			return err
		}
		exist = _exist
		return nil
	})
	return
}

func (rc *RadosConn) RbdSnapCreate(ctx context.Context, snapSpec rbd.SnapSpec) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdSnapCreate(ctx, conn, snapSpec)
	})
}

func (rc *RadosConn) RbdSnapRemove(ctx context.Context, snapSpec rbd.SnapSpec) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdSnapRemove(ctx, conn, snapSpec)
	})
}

func (rc *RadosConn) RbdSnapList(ctx context.Context, imageSpec rbd.ImageSpec) (snaps []rbd.SnapInfo, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		_snaps, err := rbd.RbdSnapList(ctx, conn, imageSpec)
		if err != nil {
			return err
		}
		snaps = _snaps
		return nil
	})
	return
}

func (rc *RadosConn) RbdList(ctx context.Context, poolSpec rbd.PoolSpec) (images []string, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		_images, err := rbd.RbdList(ctx, conn, poolSpec)
		if err != nil {
			return err
		}
		images = _images
		return nil
	})
	return
}

func (rc *RadosConn) RbdListLong(ctx context.Context, poolSpec rbd.PoolSpec) (entries []rbd.ImageListEntry, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		_entries, err := rbd.RbdListLong(ctx, conn, poolSpec)
		if err != nil {
			return err
		}
		entries = _entries
		return nil
	})
	return
}

func (rc *RadosConn) RbdStatus(ctx context.Context, imageOrSnapSpec string) (status *rbd.ImageStatus, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		status, err = rbd.RbdStatus(ctx, conn, imageOrSnapSpec)
		return err
	})
	return
}
