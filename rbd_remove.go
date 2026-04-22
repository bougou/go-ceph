package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdRemove(ctx context.Context, imageSpec ImageSpec) error {
	err := rc.Do(ctx, func() error {
		return RbdRemove(ctx, rc.conn, imageSpec)
	})
	return err
}

func RbdRemove(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec) error {
	namespaceName, poolName, imageName, err := Image(string(imageSpec))
	if err != nil {
		return err
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	if err := rbd.RemoveImage(ioctx, imageName); err != nil {
		return fmt.Errorf("failed to remove image (%s): %w", imageName, err)
	}

	return nil
}
