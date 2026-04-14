package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdChildren(ctx context.Context, imageSpec ImageSpec) ([]ImageSpec, error) {
	var children []ImageSpec = nil
	err := rc.Do(ctx, func() error {
		_children, err := RbdChildren(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		children = _children
		return nil
	})
	return children, err
}

func RbdChildren(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec) ([]ImageSpec, error) {
	if !imageSpec.Valid() {
		return nil, errInvalidImageSpec
	}

	poolName := imageSpec.Pool()
	imageName := imageSpec.Image()
	namespaceName := imageSpec.Namespace()

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	image, err := rbd.OpenImage(ioctx, imageName, rbd.NoSnapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	childrenPoolNames, childrenImageNames, err := image.ListChildren()
	if err != nil {
		return nil, fmt.Errorf("failed to list children for image (%s): %w", imageName, err)
	}

	if len(childrenPoolNames) != len(childrenImageNames) {
		return nil, fmt.Errorf("failed to list children for image (%s): %w", imageName, "number of children pool names and image names are different")
	}

	children := make([]ImageSpec, len(childrenPoolNames))
	for i, childPoolName := range childrenPoolNames {
		children[i] = NewImageSpec(childPoolName, childrenImageNames[i])
	}

	return children, nil
}

func (rc *RadosConn) RbdSnapChildren(ctx context.Context, snapSpec SnapSpec) ([]ImageSpec, error) {
	var children []ImageSpec = nil
	err := rc.Do(ctx, func() error {
		_children, err := RbdSnapChildren(ctx, rc.conn, snapSpec)
		if err != nil {
			return err
		}
		children = _children
		return nil
	})
	return children, err
}

func RbdSnapChildren(ctx context.Context, conn *rados.Conn, snapSpec SnapSpec) ([]ImageSpec, error) {
	if !snapSpec.Valid() {
		return nil, errInvalidSnapSpec
	}

	poolName := snapSpec.Pool()
	imageName := snapSpec.Image()
	namespaceName := snapSpec.Namespace()
	snapName := snapSpec.Snap()

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	image, err := rbd.OpenImage(ioctx, imageName, snapName)
	if err != nil {
		return nil, fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	childrenPoolNames, childrenImageNames, err := image.ListChildren()
	if err != nil {
		return nil, fmt.Errorf("failed to list children for image (%s): %w", imageName, err)
	}

	if len(childrenPoolNames) != len(childrenImageNames) {
		return nil, fmt.Errorf("failed to list children for image (%s): %w", imageName, "number of children pool names and image names are different")
	}

	children := make([]ImageSpec, len(childrenPoolNames))
	for i, childPoolName := range childrenPoolNames {
		children[i] = NewImageSpec(childPoolName, childrenImageNames[i])
	}

	return children, nil
}
