package rbd

import (
	"context"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

func RbdRename(ctx context.Context, conn *cephrados.Conn, srcImageSpec ImageSpec, dstImageSpec ImageSpec) error {
	srcNamespaceName, srcPoolName, srcImageName, err := Image(string(srcImageSpec))
	if err != nil {
		return err
	}
	dstNamespaceName, dstPoolName, dstImageName, err := Image(string(dstImageSpec))
	if err != nil {
		return err
	}

	if srcImageSpec.Equal(dstImageSpec) {
		return nil
	}

	if srcPoolName != dstPoolName {
		return fmt.Errorf("source pool (%s) and destination pool (%s) are different", srcPoolName, dstPoolName)
	}

	if srcNamespaceName != dstNamespaceName {
		return fmt.Errorf("source namespace (%s) and destination namespace (%s) are different", srcNamespaceName, dstNamespaceName)
	}

	ioctx, err := conn.OpenIOContext(srcPoolName)
	if err != nil {
		return fmt.Errorf("failed to open source pool (%s): %w", srcPoolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(srcNamespaceName)

	srcImage, err := cephrbd.OpenImage(ioctx, srcImageName, "")
	if err != nil {
		return fmt.Errorf("failed to open source image (%s): %w", srcImageName, err)
	}
	defer srcImage.Close()

	if err := srcImage.Rename(dstImageName); err != nil {
		return fmt.Errorf("failed to rename source image (%s) to (%s): %w", srcImageName, dstImageName, err)
	}

	return nil
}
