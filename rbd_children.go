package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdChildren(ctx context.Context, imageSpec ImageSpec) (children []ImageSpec, err error) {
	err = rc.Do(ctx, func() error {
		_children, err := RbdChildren(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		children = _children
		return nil
	})
	return
}

func RbdChildren(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec) (children []ImageSpec, err error) {
	namespaceName, poolName, imageName, err := Image(string(imageSpec))
	if err != nil {
		return
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		err = fmt.Errorf("failed to open pool (%s): %w", poolName, err)
		return
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	image, err := rbd.OpenImage(ioctx, imageName, rbd.NoSnapshot)
	if err != nil {
		err = fmt.Errorf("failed to open image (%s): %w", imageName, err)
		return
	}
	defer image.Close()

	childrenPoolNames, childrenImageNames, err := image.ListChildren()
	if err != nil {
		err = fmt.Errorf("failed to list children for image (%s): %w", imageName, err)
		return
	}

	if len(childrenPoolNames) != len(childrenImageNames) {
		err = fmt.Errorf("failed to list children for image (%s): %w", imageName, "number of children pool names and image names are different")
		return
	}

	children = make([]ImageSpec, len(childrenPoolNames))
	for i, childPoolName := range childrenPoolNames {
		children[i] = NewImageSpec(childPoolName, childrenImageNames[i])
	}

	return
}

func (rc *RadosConn) RbdSnapChildren(ctx context.Context, snapSpec SnapSpec) (children []ImageSpec, err error) {
	err = rc.Do(ctx, func() error {
		_children, err := RbdSnapChildren(ctx, rc.conn, snapSpec)
		if err != nil {
			return err
		}
		children = _children
		return nil
	})
	return
}

func RbdSnapChildren(ctx context.Context, conn *rados.Conn, snapSpec SnapSpec) (children []ImageSpec, err error) {
	namespaceName, poolName, imageName, snapName, err := Snap(string(snapSpec))
	if err != nil {
		return
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		err = fmt.Errorf("failed to open pool (%s): %w", poolName, err)
		return
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	image, err := rbd.OpenImage(ioctx, imageName, snapName)
	if err != nil {
		err = fmt.Errorf("failed to open image (%s): %w", imageName, err)
		return
	}
	defer image.Close()

	childrenPoolNames, childrenImageNames, err := image.ListChildren()
	if err != nil {
		err = fmt.Errorf("failed to list children for image (%s): %w", imageName, err)
		return
	}

	if len(childrenPoolNames) != len(childrenImageNames) {
		err = fmt.Errorf("failed to list children for image (%s): %w", imageName, "number of children pool names and image names are different")
		return
	}

	children = make([]ImageSpec, len(childrenPoolNames))
	for i, childPoolName := range childrenPoolNames {
		children[i] = NewImageSpec(childPoolName, childrenImageNames[i])
	}

	return
}
