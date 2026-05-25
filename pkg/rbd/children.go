package rbd

import (
	"context"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

func RbdChildren(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec) (children []ImageSpec, err error) {
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

	image, err := cephrbd.OpenImage(ioctx, imageName, cephrbd.NoSnapshot)
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

func RbdSnapChildren(ctx context.Context, conn *cephrados.Conn, snapSpec SnapSpec) (children []ImageSpec, err error) {
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

	image, err := cephrbd.OpenImage(ioctx, imageName, snapName)
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
