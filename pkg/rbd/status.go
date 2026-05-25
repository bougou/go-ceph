package rbd

import (
	"context"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

func RbdStatus(ctx context.Context, conn *cephrados.Conn, imageOrSnapSpec string) (watchers []cephrbd.ImageWatcher, err error) {
	namespaceName, poolName, imageName, snapshotName, err := ImageOrSnap(imageOrSnapSpec)
	if err != nil {
		return
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		err = fmt.Errorf("failed to open pool (%s): %w", poolName, err)
		return
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	snapName := snapshotName
	if snapName == "" {
		snapName = cephrbd.NoSnapshot
	}

	// Note: we use OpenImageReadOnly instead of OpenImage to avoid the need to open the image for writing.
	// OpenImage would register itself as a watcher.
	image, err := cephrbd.OpenImageReadOnly(ioctx, imageName, snapName)
	if err != nil {
		err = fmt.Errorf("failed to open image (%s): %w", imageName, err)
		return
	}
	defer image.Close()

	w, err := image.ListWatchers()
	if err != nil {
		err = fmt.Errorf("failed to get watchers: %w", err)
		return
	}

	watchers = w
	return
}
