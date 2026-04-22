package ceph

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func (rc *RadosConn) RbdStatus(ctx context.Context, imageOrSnapSpec string) (watchers []rbd.ImageWatcher, err error) {
	watchers = nil
	err = rc.Do(ctx, func() error {
		_watchers, err := RbdStatus(ctx, rc.conn, imageOrSnapSpec)
		if err != nil {
			return err
		}
		watchers = _watchers
		return nil
	})
	return watchers, err
}

func RbdStatus(ctx context.Context, conn *rados.Conn, imageOrSnapSpec string) (watchers []rbd.ImageWatcher, err error) {
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
		snapName = rbd.NoSnapshot
	}

	// Note: we use OpenImageReadOnly instead of OpenImage to avoid the need to open the image for writing.
	// OpenImage would register itself as a watcher.
	image, err := rbd.OpenImageReadOnly(ioctx, imageName, snapName)
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
