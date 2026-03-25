package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdClone(ctx context.Context, srcSnapSpec SnapSpec, dstImageSpec ImageSpec, optFns ...RbdImageOptionFn) error {
	err := rc.Do(ctx, func() error {
		return RbdClone(ctx, rc.conn, srcSnapSpec, dstImageSpec, optFns...)
	})
	return err
}

func RbdClone(ctx context.Context, conn *rados.Conn, srcSnapSpec SnapSpec, dstImageSpec ImageSpec, optFns ...RbdImageOptionFn) error {
	if !srcSnapSpec.Valid() {
		return errInvalidSnapSpec
	}

	if !dstImageSpec.Valid() {
		return errInvalidImageSpec
	}

	srcPoolName := srcSnapSpec.Pool()
	srcImageName := srcSnapSpec.Image()
	srcNamespaceName := srcSnapSpec.Namespace()
	srcSnapName := srcSnapSpec.Snap()

	dstPoolName := dstImageSpec.Pool()
	dstImageName := dstImageSpec.Image()
	dstNamespaceName := dstImageSpec.Namespace()

	if srcPoolName != dstPoolName {
		return fmt.Errorf("source pool (%s) and destination pool (%s) are different", srcPoolName, dstPoolName)
	}

	if srcNamespaceName != dstNamespaceName {
		return fmt.Errorf("source namespace (%s) and destination namespace (%s) are different", srcNamespaceName, dstNamespaceName)
	}

	srcIOCtx, err := conn.OpenIOContext(srcPoolName)
	if err != nil {
		return fmt.Errorf("failed to open source pool (%s): %w", srcPoolName, err)
	}
	defer srcIOCtx.Destroy()

	srcIOCtx.SetNamespace(srcNamespaceName)

	srcSnap, err := rbd.OpenImage(srcIOCtx, srcImageName, srcSnapName)
	if err != nil {
		return fmt.Errorf("failed to open source snapshot (%s) for image (%s): %w", srcSnapName, srcImageName, err)
	}
	defer srcSnap.Close()

	dstIOCtx, err := conn.OpenIOContext(dstPoolName)
	if err != nil {
		return fmt.Errorf("failed to open destination pool (%s): %w", dstPoolName, err)
	}
	defer dstIOCtx.Destroy()

	dstIOCtx.SetNamespace(dstNamespaceName)

	imageOpts, err := rbdImageOptionsFromFns(optFns...)
	if err != nil {
		return fmt.Errorf("failed to build image options: %w", err)
	}
	defer imageOpts.Destroy()

	if err := rbd.CloneImage(srcIOCtx, srcImageName, srcSnapName, dstIOCtx, dstImageName, imageOpts); err != nil {
		return fmt.Errorf("failed to clone image (%s) from snapshot (%s): %w", dstImageName, srcSnapSpec, err)
	}

	return nil
}
