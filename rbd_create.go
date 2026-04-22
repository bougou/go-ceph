package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdCreate(ctx context.Context, imageSpec ImageSpec, sizeBytes int64, optFns ...RbdImageOptionFn) error {
	err := rc.Do(ctx, func() error {
		return RbdCreate(ctx, rc.conn, imageSpec, sizeBytes, optFns...)
	})
	return err
}

func RbdCreate(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec, sizeBytes int64, optFns ...RbdImageOptionFn) error {
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

	if err := rbd.CreateImage(ioctx, imageName, uint64(sizeBytes), imageOpts); err != nil {
		return fmt.Errorf("failed to create image (%s): %w", imageName, err)
	}

	return nil
}
