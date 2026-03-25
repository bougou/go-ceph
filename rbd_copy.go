package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdCopy(ctx context.Context, srcImageSpec ImageSpec, dstImageSpec ImageSpec, optFns ...RbdImageOptionFn) error {
	err := rc.Do(ctx, func() error {
		return RbdCopy(ctx, rc.conn, srcImageSpec, dstImageSpec, optFns...)
	})
	return err
}

func RbdCopy(ctx context.Context, conn *rados.Conn, srcImageSpec ImageSpec, dstImageSpec ImageSpec, optFns ...RbdImageOptionFn) error {
	if !srcImageSpec.Valid() || !dstImageSpec.Valid() {
		return errInvalidImageSpec
	}

	if srcImageSpec.Equal(dstImageSpec) {
		return fmt.Errorf("source and destination image spec are the same")
	}

	srcPoolName := srcImageSpec.Pool()
	srcImageName := srcImageSpec.Image()
	srcNamespaceName := srcImageSpec.Namespace()
	dstPoolName := dstImageSpec.Pool()
	dstImageName := dstImageSpec.Image()
	dstNamespaceName := dstImageSpec.Namespace()

	srcIOCtx, err := conn.OpenIOContext(srcPoolName)
	if err != nil {
		return fmt.Errorf("failed to open source pool (%s): %w", srcPoolName, err)
	}
	defer srcIOCtx.Destroy()

	srcIOCtx.SetNamespace(srcNamespaceName)

	srcImage, err := rbd.OpenImage(srcIOCtx, srcImageName, rbd.NoSnapshot)
	if err != nil {
		return fmt.Errorf("failed to open source image (%s): %w", srcImageName, err)
	}
	defer srcImage.Close()

	tempSnapName := fmt.Sprintf("%s__temp_for_copy__", dstImageName)

	var tempSnap *rbd.Snapshot = nil

	// Snapshot existence is checked via image snapshot ID lookup.
	if _, err := srcImage.GetSnapID(tempSnapName); err != nil {
		if isErrNotFound(err) {
			tempSnap, err = srcImage.CreateSnapshot(tempSnapName)
			if err != nil {
				return fmt.Errorf("failed to create snapshot (%s): %w", tempSnapName, err)
			}
		} else {
			return fmt.Errorf("failed to query snapshot (%s): %w", tempSnapName, err)
		}
	} else {
		tempSnap = srcImage.GetSnapshot(tempSnapName)
	}

	isProtected, err := tempSnap.IsProtected()
	if err != nil {
		return fmt.Errorf("failed to check protection for snapshot (%s): %w", tempSnapName, err)
	}
	if !isProtected {
		if err := tempSnap.Protect(); err != nil {
			return fmt.Errorf("failed to protect snapshot (%s): %w", tempSnapName, err)
		}
	}

	defer func() {
		// remove the temporary snapshot
		tempSnap.Unprotect()
		tempSnap.Remove()
	}()

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

	if err := rbd.CloneFromImage(srcImage, tempSnapName, dstIOCtx, dstImageName, imageOpts); err != nil {
		return fmt.Errorf("failed to clone destination image (%s) from snapshot (%s): %w", dstImageName, tempSnapName, err)
	}

	// flatten the destination image
	dstImage, err := rbd.OpenImage(dstIOCtx, dstImageName, "")
	if err != nil {
		return fmt.Errorf("failed to open destination image (%s): %w", dstImageName, err)
	}
	defer dstImage.Close()

	if err := dstImage.Flatten(); err != nil {
		return fmt.Errorf("failed to flatten destination image (%s): %w", dstImageName, err)
	}

	return nil
}
