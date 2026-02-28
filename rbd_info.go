package ceph

import (
	"context"
	"errors"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

// RbdInfo retrieves detailed information about an RBD image.
// This method already handled the image ErrNotFound error.
// If the image does not exist, it returns nil, nil.
func (rc *RadosConn) RbdInfo(ctx context.Context, imageSpec ImageSpec) (*RbdImageInfo, error) {
	var info *RbdImageInfo = nil

	err := rc.Do(ctx, func() error {
		_info, err := RbdInfo(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		info = _info
		return nil
	})

	return info, err
}

func RbdInfo(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec) (*RbdImageInfo, error) {
	poolName := imageSpec.Pool()
	imageName := imageSpec.Image()
	namespaceName := imageSpec.Namespace()

	// Open IOContext for the pool
	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	image, err := rbd.OpenImage(ioctx, imageName, rbd.NoSnapshot)
	if err != nil {
		if errors.Is(err, rbd.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	return ConvertRbdImageToRbdInfo(image)
}
