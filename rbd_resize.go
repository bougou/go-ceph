package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdResize(ctx context.Context, imageSpec ImageSpec, sizeBytes uint64) error {
	err := rc.Do(ctx, func() error {
		return RbdResize(ctx, rc.conn, imageSpec, sizeBytes)
	})
	return err
}

func RbdResize(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec, sizeBytes uint64) error {
	poolName := imageSpec.Pool()
	imageName := imageSpec.Image()
	namespaceName := imageSpec.Namespace()

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

	if err := image.Resize(sizeBytes); err != nil {
		return fmt.Errorf("failed to resize image (%s): %w", imageName, err)
	}

	return nil
}
