package rbd

import (
	"context"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

// RbdParents returns all parent images of the given image, walking up the clone
// chain from the nearest parent to the root. An image with no parent returns nil.
func RbdParents(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec) (parents []ImageSpec, err error) {
	namespaceName, poolName, imageName, err := Image(string(imageSpec))
	if err != nil {
		return
	}

	return rbdListParents(ctx, conn, namespaceName, poolName, imageName, "")
}

// RbdSnapParents returns all parent images of the given snapshot, walking up the
// clone chain from the nearest parent to the root. A snapshot with no parent returns nil.
func RbdSnapParents(ctx context.Context, conn *cephrados.Conn, snapSpec SnapSpec) (parents []ImageSpec, err error) {
	namespaceName, poolName, imageName, snapName, err := Snap(string(snapSpec))
	if err != nil {
		return
	}

	return rbdListParents(ctx, conn, namespaceName, poolName, imageName, snapName)
}

func rbdListParents(ctx context.Context, conn *cephrados.Conn, namespaceName, poolName, imageName, snapName string) (parents []ImageSpec, err error) {
	currentNamespace := namespaceName
	currentPool := poolName
	currentImage := imageName
	currentSnap := snapName

	seen := make(map[string]struct{})

	for {
		parentSpec, parentErr := rbdGetParent(ctx, conn, currentNamespace, currentPool, currentImage, currentSnap)
		if parentErr != nil {
			if isErrNotFound(parentErr) {
				return parents, nil
			}
			err = parentErr
			return
		}

		parentKey := string(parentSpec)
		if _, ok := seen[parentKey]; ok {
			err = fmt.Errorf("parent chain cycle detected at %s", parentSpec)
			return
		}
		seen[parentKey] = struct{}{}
		parents = append(parents, parentSpec)

		currentNamespace = parentSpec.Namespace()
		currentPool = parentSpec.Pool()
		currentImage = parentSpec.Image()
		currentSnap = ""
	}
}

func rbdGetParent(ctx context.Context, conn *cephrados.Conn, namespaceName, poolName, imageName, snapName string) (parentSpec ImageSpec, err error) {
	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		err = fmt.Errorf("failed to open pool (%s): %w", poolName, err)
		return
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	var image *cephrbd.Image
	if snapName == "" {
		image, err = cephrbd.OpenImage(ioctx, imageName, cephrbd.NoSnapshot)
	} else {
		image, err = cephrbd.OpenImage(ioctx, imageName, snapName)
	}
	if err != nil {
		err = fmt.Errorf("failed to open image (%s): %w", imageName, err)
		return
	}
	defer image.Close()

	parentInfo, err := image.GetParent()
	if err != nil {
		if isErrNotFound(err) {
			return "", err
		}
		err = fmt.Errorf("failed to get parent for image (%s): %w", imageName, err)
		return
	}

	parentSpec = NewImageSpecWithNamespace(
		parentInfo.Image.PoolName,
		parentInfo.Image.PoolNamespace,
		parentInfo.Image.ImageName,
	)
	return
}
