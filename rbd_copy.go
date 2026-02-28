package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdCopy(ctx context.Context, srcImageSpec ImageSpec, dstImageSpec ImageSpec) error {
	err := rc.Do(ctx, func() error {
		return RbdCopy(ctx, rc.conn, srcImageSpec, dstImageSpec)
	})
	return err
}

func RbdCopy(ctx context.Context, conn *rados.Conn, srcImageSpec ImageSpec, dstImageSpec ImageSpec) error {
	srcPoolName := srcImageSpec.Pool()
	srcImageName := srcImageSpec.Image()
	srcNamespaceName := srcImageSpec.Namespace()
	dstPoolName := dstImageSpec.Pool()
	dstImageName := dstImageSpec.Image()
	dstNamespaceName := dstImageSpec.Namespace()

	if srcImageName == "" || dstImageName == "" {
		return fmt.Errorf("source or destination image name is empty")
	}

	if srcImageName == dstImageName {
		return fmt.Errorf("source and destination image name are the same")
	}

	srcIOCtx, err := conn.OpenIOContext(srcPoolName)
	if err != nil {
		return fmt.Errorf("failed to open source pool (%s): %w", srcPoolName, err)
	}
	defer srcIOCtx.Destroy()

	srcIOCtx.SetNamespace(srcNamespaceName)

	srcImage, err := rbd.OpenImage(srcIOCtx, srcImageName, "")
	if err != nil {
		return fmt.Errorf("failed to open source image (%s): %w", srcImageName, err)
	}
	defer srcImage.Close()

	tempSnapName := fmt.Sprintf("%s__temp_for_copy__", dstImageName)
	tempSnapSpec := SnapSpec(fmt.Sprintf("%s@%s", srcImageName, tempSnapName))

	snapExists, err := snapExist(srcIOCtx, srcImageName, tempSnapName)
	if err != nil {
		return fmt.Errorf("failed to check if snapshot (%s) exists: %w", tempSnapSpec, err)
	}
	if snapExists {
		return fmt.Errorf("snapshot (%s) already exists", tempSnapSpec)
	}

	srcSnap, err := srcImage.CreateSnapshot(tempSnapName)
	if err != nil {
		return fmt.Errorf("failed to create snapshot (%s): %w", tempSnapName, err)
	}

	if err := srcSnap.Protect(); err != nil {
		srcSnap.Unprotect()
		srcSnap.Remove()
		return fmt.Errorf("failed to protect snapshot (%s): %w", tempSnapName, err)
	}

	dstIOCtx, err := conn.OpenIOContext(dstPoolName)
	if err != nil {
		return fmt.Errorf("failed to open destination pool (%s): %w", dstPoolName, err)
	}
	defer dstIOCtx.Destroy()

	dstIOCtx.SetNamespace(dstNamespaceName)

	opts := rbd.NewRbdImageOptions()
	opts.SetUint64(rbd.ImageOption(rbd.ImageOptionFeatures), DefaultImageFeatures)
	opts.SetUint64(rbd.ImageOption(rbd.ImageOptionOrder), DefaultImageOrder)
	rbd.CloneFromImage(srcImage, tempSnapName, dstIOCtx, dstImageName, opts)

	return nil
}
