package rbd

import (
	"context"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

func RbdRemove(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec) error {
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

	if err := cephrbd.RemoveImage(ioctx, imageName); err != nil {
		return fmt.Errorf("failed to remove image (%s): %w", imageName, err)
	}

	return nil
}
