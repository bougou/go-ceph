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

// RbdForceRemove removes an image after severing clone dependencies.
//
// This is a blocking (synchronous) operation: the caller waits until every
// child flatten, snapshot removal, and the final image delete have finished.
// Child flatten uses librbd image.Flatten() (same as RbdFlatten), not a
// background "rbd task add flatten". Runtime grows with child image size and
// the number of clones/snapshots.
//
// Steps:
//  1. For every snapshot, flatten live clone children (children are kept as
//     independent images) and permanently delete trash-linked children.
//  2. Unprotect and remove all snapshots on the image.
//  3. Remove the image itself.
//
// Child images are not deleted unless they are already in trash.
func RbdForceRemove(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec) error {
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

	image, err := cephrbd.OpenImage(ioctx, imageName, cephrbd.NoSnapshot)
	if err != nil {
		return fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}

	snaps, err := image.GetSnapshotNames()
	if err != nil {
		_ = image.Close()
		return fmt.Errorf("failed to list snapshots for image (%s): %w", imageName, err)
	}

	if err := flattenCloneChildren(ctx, conn, ioctx, imageName, snaps); err != nil {
		_ = image.Close()
		return err
	}

	if err := removeAllSnapshots(image, imageName, snaps); err != nil {
		_ = image.Close()
		return err
	}

	if err := image.Close(); err != nil {
		return fmt.Errorf("failed to close image (%s): %w", imageName, err)
	}

	if err := cephrbd.RemoveImage(ioctx, imageName); err != nil {
		return fmt.Errorf("failed to remove image (%s): %w", imageName, err)
	}

	return nil
}

func flattenCloneChildren(ctx context.Context, conn *cephrados.Conn, parentIOCtx *cephrados.IOContext, imageName string, snaps []cephrbd.SnapInfo) error {
	seen := make(map[string]struct{})

	for _, snap := range snaps {
		if err := ctx.Err(); err != nil {
			return err
		}

		snapImage, err := cephrbd.OpenImage(parentIOCtx, imageName, snap.Name)
		if err != nil {
			return fmt.Errorf("failed to open snapshot (%s) for image (%s): %w", snap.Name, imageName, err)
		}

		children, err := snapImage.ListChildrenAttributes()
		_ = snapImage.Close()
		if err != nil {
			return fmt.Errorf("failed to list children for snapshot (%s) of image (%s): %w", snap.Name, imageName, err)
		}

		for _, child := range children {
			key := child.PoolName + "/" + child.PoolNamespace + "/" + child.ImageID
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}

			if err := ctx.Err(); err != nil {
				return err
			}

			if child.Trash {
				if err := permanentlyRemoveTrashChild(conn, child); err != nil {
					return err
				}
				continue
			}

			childSpec := NewImageSpecWithNamespace(child.PoolName, child.PoolNamespace, child.ImageName)
			if err := RbdFlatten(ctx, conn, childSpec); err != nil {
				return fmt.Errorf("failed to flatten child image (%s) of snapshot (%s): %w", childSpec, snap.Name, err)
			}
		}
	}

	return nil
}

func permanentlyRemoveTrashChild(conn *cephrados.Conn, child cephrbd.ImageSpec) error {
	ioctx, err := conn.OpenIOContext(child.PoolName)
	if err != nil {
		return fmt.Errorf("failed to open pool (%s) for trash child (%s): %w", child.PoolName, child.ImageName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(child.PoolNamespace)

	if err := cephrbd.TrashRemove(ioctx, child.ImageID, true); err != nil {
		return fmt.Errorf("failed to permanently remove trash child image (%s id=%s): %w", child.ImageName, child.ImageID, err)
	}
	return nil
}

func removeAllSnapshots(image *cephrbd.Image, imageName string, snaps []cephrbd.SnapInfo) error {
	for _, snapInfo := range snaps {
		snap := image.GetSnapshot(snapInfo.Name)

		isProtected, err := snap.IsProtected()
		if err != nil {
			return fmt.Errorf("failed to check if snapshot (%s) is protected for image (%s): %w", snapInfo.Name, imageName, err)
		}
		if isProtected {
			if err := snap.Unprotect(); err != nil {
				return fmt.Errorf("failed to unprotect snapshot (%s) for image (%s): %w", snapInfo.Name, imageName, err)
			}
		}

		if err := snap.Remove(); err != nil {
			return fmt.Errorf("failed to remove snapshot (%s) for image (%s): %w", snapInfo.Name, imageName, err)
		}
	}
	return nil
}
