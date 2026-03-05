package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdSnapExist(ctx context.Context, snapSpec SnapSpec) (bool, error) {
	var exist bool = false
	err := rc.Do(ctx, func() error {
		_exist, err := RbdSnapExist(ctx, rc.conn, snapSpec)
		if err != nil {
			return err
		}
		exist = _exist
		return nil
	})
	return exist, err
}

func RbdSnapExist(ctx context.Context, conn *rados.Conn, snapSpec SnapSpec) (bool, error) {
	if !snapSpec.Valid() {
		return false, errInvalidSnapSpec
	}

	poolName := snapSpec.Pool()
	namespaceName := snapSpec.Namespace()
	imageName := snapSpec.Image()
	snapName := snapSpec.Snap()

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return false, fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	image, err := rbd.OpenImage(ioctx, imageName, snapName)
	if err != nil {
		if isErrNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to open image %s: %w", imageName, err)
	}
	defer image.Close()

	return true, nil
}

func (rc *RadosConn) RbdSnapCreate(ctx context.Context, snapSpec SnapSpec) error {
	err := rc.Do(ctx, func() error {
		return RbdSnapCreate(ctx, rc.conn, snapSpec)
	})
	return err
}

func RbdSnapCreate(ctx context.Context, conn *rados.Conn, snapSpec SnapSpec) error {
	if !snapSpec.Valid() {
		return errInvalidSnapSpec
	}

	poolName := snapSpec.Pool()
	imageName := snapSpec.Image()
	namespaceName := snapSpec.Namespace()
	snapName := snapSpec.Snap()

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	// Open base image before operating on its snapshot metadata.
	image, err := rbd.OpenImage(ioctx, imageName, rbd.NoSnapshot)
	if err != nil {
		return fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	rs, err := image.CreateSnapshot(snapName)
	if err != nil {
		return fmt.Errorf("failed to create snapshot (%s) for image (%s): %w", snapName, imageName, err)
	}

	if err := rs.Protect(); err != nil {
		return fmt.Errorf("failed to protect snapshot (%s) for image (%s): %w", snapName, imageName, err)
	}

	return nil
}

func (rc *RadosConn) RbdSnapRemove(ctx context.Context, snapSpec SnapSpec) error {
	err := rc.Do(ctx, func() error {
		return RbdSnapRemove(ctx, rc.conn, snapSpec)
	})
	return err
}

func RbdSnapRemove(ctx context.Context, conn *rados.Conn, snapSpec SnapSpec) error {
	if !snapSpec.Valid() {
		return errInvalidSnapSpec
	}

	poolName := snapSpec.Pool()
	imageName := snapSpec.Image()
	namespaceName := snapSpec.Namespace()
	snapName := snapSpec.Snap()

	if snapName == "" || imageName == "" {
		return fmt.Errorf("snapshot or image name is empty")
	}

	if poolName == "" {
		return fmt.Errorf("pool name is empty")
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	image, err := rbd.OpenImage(ioctx, imageName, snapName)
	if err != nil {
		return fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	snap := image.GetSnapshot(snapName)

	isProtected, err := snap.IsProtected()
	if err != nil {
		return fmt.Errorf("failed to check if snapshot (%s) is protected for image (%s): %w", snapName, imageName, err)
	}
	if isProtected {
		if err := snap.Unprotect(); err != nil {
			return fmt.Errorf("failed to unprotect snapshot (%s) for image (%s): %w", snapName, imageName, err)
		}
	}

	if err := snap.Remove(); err != nil {
		return fmt.Errorf("failed to remove snapshot (%s) for image (%s): %w", snapName, imageName, err)
	}

	return nil
}
