package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdRename(ctx context.Context, srcImageSpec ImageSpec, dstImageSpec ImageSpec) error {
	err := rc.Do(ctx, func() error {
		return RbdRename(ctx, rc.conn, srcImageSpec, dstImageSpec)
	})
	return err
}

func RbdRename(ctx context.Context, conn *rados.Conn, srcImageSpec ImageSpec, dstImageSpec ImageSpec) error {
	srcPoolName := srcImageSpec.Pool()
	srcImageName := srcImageSpec.Image()
	srcNamespaceName := srcImageSpec.Namespace()
	dstPoolName := dstImageSpec.Pool()
	dstImageName := dstImageSpec.Image()
	dstNamespaceName := dstImageSpec.Namespace()

	if srcImageName == "" || dstImageName == "" {
		return fmt.Errorf("source or destination image name is empty")
	}

	if srcPoolName != dstPoolName {
		return fmt.Errorf("source pool (%s) and destination pool (%s) are different", srcPoolName, dstPoolName)
	}

	if srcNamespaceName != dstNamespaceName {
		return fmt.Errorf("source namespace (%s) and destination namespace (%s) are different", srcNamespaceName, dstNamespaceName)
	}

	if srcImageName == dstImageName {
		return nil
	}

	ioctx, err := conn.OpenIOContext(srcPoolName)
	if err != nil {
		return fmt.Errorf("failed to open source pool (%s): %w", srcPoolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(srcNamespaceName)

	srcImage, err := rbd.OpenImage(ioctx, srcImageName, "")
	if err != nil {
		return fmt.Errorf("failed to open source image (%s): %w", srcImageName, err)
	}
	defer srcImage.Close()

	if err := srcImage.Rename(dstImageName); err != nil {
		return fmt.Errorf("failed to rename source image (%s) to (%s): %w", srcImageName, dstImageName, err)
	}

	return nil
}
