package rbd

import (
	"context"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

func RbdCreate(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec, sizeBytes int64, optFns ...RbdImageOptionFn) error {
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

	imageOpts, err := rbdImageOptionsFromFns(optFns...)
	if err != nil {
		return fmt.Errorf("failed to build image options: %w", err)
	}
	defer imageOpts.Destroy()

	if err := cephrbd.CreateImage(ioctx, imageName, uint64(sizeBytes), imageOpts); err != nil {
		return fmt.Errorf("failed to create image (%s): %w", imageName, err)
	}

	return nil
}
