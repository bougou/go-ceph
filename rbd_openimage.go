package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

// RbdOpenImage opens an RBD image and returns it.
// You should close the image after using it.
func (rc *RadosConn) RbdOpenImage(ctx context.Context, imageSpec ImageSpec) (image *rbd.Image, err error) {
	err = rc.Do(ctx, func() error {
		_image, err := RbdOpenImage(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		image = _image
		return nil
	})
	return
}

// RbdOpenImage opens an RBD image and returns it.
// You should close the image after using it.
func RbdOpenImage(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec) (image *rbd.Image, err error) {
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

	image, err = rbd.OpenImage(ioctx, imageName, "")
	if err != nil {
		err = fmt.Errorf("failed to open image (%s): %w", imageName, err)
		return
	}

	return
}
