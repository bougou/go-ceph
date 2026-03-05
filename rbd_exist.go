package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdExist(ctx context.Context, imageSpec ImageSpec) (bool, error) {
	var exist bool = false
	err := rc.Do(ctx, func() error {
		_exist, err := RbdExist(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		exist = _exist
		return nil
	})
	return exist, err
}

func RbdExist(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec) (bool, error) {
	if !imageSpec.Valid() {
		return false, errInvalidImageSpec
	}

	poolName := imageSpec.Pool()
	imageName := imageSpec.Image()
	namespaceName := imageSpec.Namespace()

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return false, fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	image, err := rbd.OpenImageReadOnly(ioctx, imageName, rbd.NoSnapshot)
	if err != nil {
		if isErrNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	return true, nil
}
