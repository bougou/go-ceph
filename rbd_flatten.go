package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdFlatten(ctx context.Context, imageSpec ImageSpec) error {
	err := rc.Do(ctx, func() error {
		return RbdFlatten(ctx, rc.conn, imageSpec)
	})
	return err
}

func RbdFlatten(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec) error {
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

	image, err := rbd.OpenImage(ioctx, imageName, "")
	if err != nil {
		return fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	if err := image.Flatten(); err != nil {
		return fmt.Errorf("failed to flatten image (%s): %w", imageName, err)
	}

	return nil
}
