package rbd

import (
	"context"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

func RbdFlatten(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec) error {
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

	return flattenImage(ioctx, imageName)
}

// flattenImage synchronously flattens a cloned image via librbd.
func flattenImage(dstIOCtx *cephrados.IOContext, imageName string) error {
	image, err := cephrbd.OpenImage(dstIOCtx, imageName, cephrbd.NoSnapshot)
	if err != nil {
		return fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	if err := image.Flatten(); err != nil {
		return fmt.Errorf("failed to flatten image (%s): %w", imageName, err)
	}
	return nil
}
