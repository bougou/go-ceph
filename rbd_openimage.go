package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

// RbdOpenImage opens an RBD image and returns it.
// You should close the image after using it.
func (rc *RadosConn) RbdOpenImage(ctx context.Context, imageSpec ImageSpec) (*rbd.Image, error) {
	var image *rbd.Image = nil
	err := rc.Do(ctx, func() error {
		_image, err := RbdOpenImage(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		image = _image
		return nil
	})

	return image, err
}

// RbdOpenImage opens an RBD image and returns it.
// You should close the image after using it.
func RbdOpenImage(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec) (*rbd.Image, error) {
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

	image, err := rbd.OpenImage(ioctx, imageName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}

	return image, nil
}
