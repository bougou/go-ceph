package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdCreate(ctx context.Context, imageSpec ImageSpec, sizeBytes int64, features uint64, order int) error {
	err := rc.Do(ctx, func() error {
		return RbdCreate(ctx, rc.conn, imageSpec, sizeBytes, features, order)
	})
	return err
}

func RbdCreate(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec, sizeBytes int64, features uint64, order int) error {
	poolName := imageSpec.Pool()
	imageName := imageSpec.Image()
	namespaceName := imageSpec.Namespace()

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	image, err := rbd.Create2(ioctx, imageName, uint64(sizeBytes), features, int(order))
	if err != nil {
		return fmt.Errorf("failed to create image (%s): %w", imageName, err)
	}
	defer image.Close()

	return nil
}
