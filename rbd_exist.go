package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdExist(ctx context.Context, imageSpec ImageSpec) (exist bool, err error) {
	err = rc.Do(ctx, func() error {
		_exist, err := RbdExist(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		exist = _exist
		return nil
	})
	return
}

func RbdExist(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec) (exist bool, err error) {
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

	image, err := rbd.OpenImageReadOnly(ioctx, imageName, rbd.NoSnapshot)
	if err != nil {
		if isErrNotFound(err) {
			err = nil
			return
		}
		err = fmt.Errorf("failed to open image (%s): %w", imageName, err)
		return
	}
	defer image.Close()

	exist = true
	return
}
