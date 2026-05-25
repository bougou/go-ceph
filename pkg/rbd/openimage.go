package rbd

import (
	"context"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

// RbdOpenImage opens an RBD image and returns it.
// You should close the image after using it.
func RbdOpenImage(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec) (image *cephrbd.Image, err error) {
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

	image, err = cephrbd.OpenImage(ioctx, imageName, "")
	if err != nil {
		err = fmt.Errorf("failed to open image (%s): %w", imageName, err)
		return
	}

	return
}
