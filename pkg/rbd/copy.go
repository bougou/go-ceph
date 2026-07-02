package rbd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

func tempCopySnapName(dstImageName string) (string, error) {
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", fmt.Errorf("failed to generate temporary snapshot name: %w", err)
	}
	return fmt.Sprintf("%s__temp__%s", dstImageName, hex.EncodeToString(suffix[:])), nil
}

func tempCopyDestImageName(dstImageName string) (string, error) {
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", fmt.Errorf("failed to generate temporary destination image name: %w", err)
	}
	return fmt.Sprintf("%s__temp__%s", dstImageName, hex.EncodeToString(suffix[:])), nil
}

func assertDestImageAvailable(ctx context.Context, conn *cephrados.Conn, dstNamespaceName, dstPoolName, dstImageName string) (ImageSpec, error) {
	dstImageSpec := NewImageSpecWithNamespace(dstPoolName, dstNamespaceName, dstImageName)
	exist, err := RbdExist(ctx, conn, dstImageSpec)
	if err != nil {
		return "", fmt.Errorf("failed to check destination image (%s): %w", dstImageName, err)
	}
	if exist {
		return "", fmt.Errorf("destination image already exists: %s", dstImageSpec)
	}
	return dstImageSpec, nil
}

func removeTempDestImage(dstIOCtx *cephrados.IOContext, tempDestImageName string) {
	if removeErr := cephrbd.RemoveImage(dstIOCtx, tempDestImageName); removeErr != nil && !isErrNotFound(removeErr) {
		_ = removeErr
	}
}

func renameTempDestImage(dstIOCtx *cephrados.IOContext, tempDestImageName, dstImageName string) error {
	tempImage, err := cephrbd.OpenImage(dstIOCtx, tempDestImageName, cephrbd.NoSnapshot)
	if err != nil {
		return fmt.Errorf("failed to open temporary destination image (%s): %w", tempDestImageName, err)
	}
	defer tempImage.Close()

	if err := tempImage.Rename(dstImageName); err != nil {
		return fmt.Errorf("failed to rename temporary destination image (%s) to (%s): %w", tempDestImageName, dstImageName, err)
	}
	return nil
}

// withStagedDestImage copies into a temporary destination image, then renames it
// to dstImageName on success. The temporary image is removed when fn returns an error.
func withStagedDestImage(ctx context.Context, conn *cephrados.Conn, dstNamespaceName, dstPoolName, dstImageName string, fn func(dstIOCtx *cephrados.IOContext, tempDestImageName string) error) error {
	if _, err := assertDestImageAvailable(ctx, conn, dstNamespaceName, dstPoolName, dstImageName); err != nil {
		return err
	}

	tempDestImageName, err := tempCopyDestImageName(dstImageName)
	if err != nil {
		return err
	}

	dstIOCtx, err := conn.OpenIOContext(dstPoolName)
	if err != nil {
		return fmt.Errorf("failed to open destination pool (%s): %w", dstPoolName, err)
	}
	defer dstIOCtx.Destroy()

	dstIOCtx.SetNamespace(dstNamespaceName)

	finalized := false
	defer func() {
		if !finalized {
			removeTempDestImage(dstIOCtx, tempDestImageName)
		}
	}()

	if err := fn(dstIOCtx, tempDestImageName); err != nil {
		return err
	}

	if err := renameTempDestImage(dstIOCtx, tempDestImageName, dstImageName); err != nil {
		return err
	}
	finalized = true
	return nil
}

// withTempCopySnapshot creates a temporary source snapshot for copy, protects it,
// then removes and unprotects it after fn returns.
func withTempCopySnapshot(srcImage *cephrbd.Image, dstImageName string, fn func(tempSnapName string) error) error {
	tempSnapName, err := tempCopySnapName(dstImageName)
	if err != nil {
		return err
	}

	tempSnap, err := srcImage.CreateSnapshot(tempSnapName)
	if err != nil {
		return fmt.Errorf("failed to create snapshot (%s): %w", tempSnapName, err)
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
		_ = tempSnap.Unprotect()
		_ = tempSnap.Remove()
	}()

	return fn(tempSnapName)
}

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

// RbdCopy copies a source image to a new independent image with point-in-time consistency.
// It creates a temporary snapshot on the source, clones from that snapshot, flattens the
// clone, then removes the temporary snapshot.
func RbdCopy(ctx context.Context, conn *cephrados.Conn, srcImageSpec ImageSpec, dstImageSpec ImageSpec, optFns ...RbdImageOptionFn) error {
	srcNamespaceName, srcPoolName, srcImageName, err := Image(string(srcImageSpec))
	if err != nil {
		return err
	}
	dstNamespaceName, dstPoolName, dstImageName, err := Image(string(dstImageSpec))
	if err != nil {
		return err
	}

	if srcImageSpec.Equal(dstImageSpec) {
		return fmt.Errorf("source and destination image spec are the same")
	}

	srcIOCtx, err := conn.OpenIOContext(srcPoolName)
	if err != nil {
		return fmt.Errorf("failed to open source pool (%s): %w", srcPoolName, err)
	}
	defer srcIOCtx.Destroy()

	srcIOCtx.SetNamespace(srcNamespaceName)

	srcImage, err := cephrbd.OpenImage(srcIOCtx, srcImageName, cephrbd.NoSnapshot)
	if err != nil {
		return fmt.Errorf("failed to open source image (%s): %w", srcImageName, err)
	}
	defer srcImage.Close()

	imageOpts, err := rbdImageOptionsFromFns(optFns...)
	if err != nil {
		return fmt.Errorf("failed to build image options: %w", err)
	}
	defer imageOpts.Destroy()

	return withTempCopySnapshot(srcImage, dstImageName, func(tempSnapName string) error {
		return withStagedDestImage(ctx, conn, dstNamespaceName, dstPoolName, dstImageName, func(dstIOCtx *cephrados.IOContext, tempDestImageName string) error {
			if err := cephrbd.CloneFromImage(srcImage, tempSnapName, dstIOCtx, tempDestImageName, imageOpts); err != nil {
				return fmt.Errorf("failed to clone temporary destination image (%s) from snapshot (%s): %w", tempDestImageName, tempSnapName, err)
			}
			return flattenImage(dstIOCtx, tempDestImageName)
		})
	})
}

// RbdCopySnap copies an existing snapshot to a new independent image by cloning from
// the snapshot and flattening the clone.
func RbdCopySnap(ctx context.Context, conn *cephrados.Conn, srcSnapSpec SnapSpec, dstImageSpec ImageSpec, optFns ...RbdImageOptionFn) error {
	if !srcSnapSpec.Valid() {
		return fmt.Errorf("invalid source snapshot spec: %s", srcSnapSpec)
	}

	srcNamespaceName, srcPoolName, srcImageName, srcSnapName, err := Snap(string(srcSnapSpec))
	if err != nil {
		return err
	}
	dstNamespaceName, dstPoolName, dstImageName, err := Image(string(dstImageSpec))
	if err != nil {
		return err
	}

	srcIOCtx, err := conn.OpenIOContext(srcPoolName)
	if err != nil {
		return fmt.Errorf("failed to open source pool (%s): %w", srcPoolName, err)
	}
	defer srcIOCtx.Destroy()

	srcIOCtx.SetNamespace(srcNamespaceName)

	srcImage, err := cephrbd.OpenImage(srcIOCtx, srcImageName, cephrbd.NoSnapshot)
	if err != nil {
		return fmt.Errorf("failed to open source image (%s): %w", srcImageName, err)
	}
	defer srcImage.Close()

	snap := srcImage.GetSnapshot(srcSnapName)

	isProtected, err := snap.IsProtected()
	if err != nil {
		return fmt.Errorf("failed to check protection for snapshot (%s): %w", srcSnapName, err)
	}
	protectedByUs := false
	if !isProtected {
		if err := snap.Protect(); err != nil {
			return fmt.Errorf("failed to protect snapshot (%s): %w", srcSnapName, err)
		}
		protectedByUs = true
	}
	if protectedByUs {
		defer func() {
			_ = snap.Unprotect()
		}()
	}

	imageOpts, err := rbdImageOptionsFromFns(optFns...)
	if err != nil {
		return fmt.Errorf("failed to build image options: %w", err)
	}
	defer imageOpts.Destroy()

	return withStagedDestImage(ctx, conn, dstNamespaceName, dstPoolName, dstImageName, func(dstIOCtx *cephrados.IOContext, tempDestImageName string) error {
		if err := cephrbd.CloneFromImage(srcImage, srcSnapName, dstIOCtx, tempDestImageName, imageOpts); err != nil {
			return fmt.Errorf("failed to clone temporary destination image (%s) from snapshot (%s): %w", tempDestImageName, srcSnapName, err)
		}
		return flattenImage(dstIOCtx, tempDestImageName)
	})
}

// RbdCopyUnsafe copies the head of a source image to a new image using librbd copy
// (same as "rbd cp pool/src pool/dst"). The copy may be inconsistent if the source
// image is being written while the copy is in progress.
func RbdCopyUnsafe(ctx context.Context, conn *cephrados.Conn, srcImageSpec ImageSpec, dstImageSpec ImageSpec) error {
	if srcImageSpec.Equal(dstImageSpec) {
		return fmt.Errorf("source and destination image spec are the same")
	}

	srcNamespaceName, srcPoolName, srcImageName, err := Image(string(srcImageSpec))
	if err != nil {
		return err
	}
	dstNamespaceName, dstPoolName, dstImageName, err := Image(string(dstImageSpec))
	if err != nil {
		return err
	}

	srcIOCtx, err := conn.OpenIOContext(srcPoolName)
	if err != nil {
		return fmt.Errorf("failed to open source pool (%s): %w", srcPoolName, err)
	}
	defer srcIOCtx.Destroy()

	srcIOCtx.SetNamespace(srcNamespaceName)

	srcImage, err := cephrbd.OpenImage(srcIOCtx, srcImageName, cephrbd.NoSnapshot)
	if err != nil {
		return fmt.Errorf("failed to open source image (%s): %w", srcImageName, err)
	}
	defer srcImage.Close()

	return withStagedDestImage(ctx, conn, dstNamespaceName, dstPoolName, dstImageName, func(dstIOCtx *cephrados.IOContext, tempDestImageName string) error {
		if err := srcImage.Copy(dstIOCtx, tempDestImageName); err != nil {
			return fmt.Errorf("failed to copy to temporary destination image (%s): %w", tempDestImageName, err)
		}
		return nil
	})
}
