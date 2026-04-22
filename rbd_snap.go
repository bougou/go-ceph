package ceph

import (
	"context"
	"fmt"
	"time"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

// RbdSnapInfo is an enriched snapshot model for API/CLI output.
// NOTE: official go-ceph currently exposes snapshot ID/Name/Size only.
// Snapshot creation timestamp is kept as an optional field for future support.
type SnapInfo struct {
	// ID is the internal snapshot ID
	ID uint64 `json:"id"`

	// Name is the name of the snapshot
	Name string `json:"name"`

	// Size is the size of the snapshot in bytes
	Size uint64 `json:"size"`

	// Protected is true if the snapshot is protected
	Protected bool `json:"protected"`

	// Timestamp is the creation timestamp of the snapshot
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// SizeHuman returns a human-readable snapshot size (e.g., "20 GiB").
func (s SnapInfo) SizeHuman() string {
	return sizeHuman(s.Size, 0)
}

func (rc *RadosConn) RbdSnapExist(ctx context.Context, snapSpec SnapSpec) (exist bool, err error) {
	err = rc.Do(ctx, func() error {
		_exist, err := RbdSnapExist(ctx, rc.conn, snapSpec)
		if err != nil {
			return err
		}
		exist = _exist
		return nil
	})
	return
}

func RbdSnapExist(ctx context.Context, conn *rados.Conn, snapSpec SnapSpec) (exist bool, err error) {
	namespaceName, poolName, imageName, snapName, err := Snap(string(snapSpec))
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

	image, err := rbd.OpenImage(ioctx, imageName, snapName)
	if err != nil {
		if isErrNotFound(err) {
			err = nil
			return
		}
		err = fmt.Errorf("failed to open image %s: %w", imageName, err)
		return
	}
	defer image.Close()

	exist = true
	return
}

func (rc *RadosConn) RbdSnapCreate(ctx context.Context, snapSpec SnapSpec) error {
	err := rc.Do(ctx, func() error {
		return RbdSnapCreate(ctx, rc.conn, snapSpec)
	})
	return err
}

func RbdSnapCreate(ctx context.Context, conn *rados.Conn, snapSpec SnapSpec) error {
	namespaceName, poolName, imageName, snapName, err := Snap(string(snapSpec))
	if err != nil {
		return err
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	// Open base image before operating on its snapshot metadata.
	image, err := rbd.OpenImage(ioctx, imageName, rbd.NoSnapshot)
	if err != nil {
		return fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	rs, err := image.CreateSnapshot(snapName)
	if err != nil {
		return fmt.Errorf("failed to create snapshot (%s) for image (%s): %w", snapName, imageName, err)
	}

	if err := rs.Protect(); err != nil {
		return fmt.Errorf("failed to protect snapshot (%s) for image (%s): %w", snapName, imageName, err)
	}

	return nil
}

func (rc *RadosConn) RbdSnapRemove(ctx context.Context, snapSpec SnapSpec) error {
	err := rc.Do(ctx, func() error {
		return RbdSnapRemove(ctx, rc.conn, snapSpec)
	})
	return err
}

func RbdSnapRemove(ctx context.Context, conn *rados.Conn, snapSpec SnapSpec) error {
	namespaceName, poolName, imageName, snapName, err := Snap(string(snapSpec))
	if err != nil {
		return err
	}

	if snapName == "" || imageName == "" {
		return fmt.Errorf("snapshot or image name is empty")
	}

	if poolName == "" {
		return fmt.Errorf("pool name is empty")
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	image, err := rbd.OpenImage(ioctx, imageName, snapName)
	if err != nil {
		return fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	snap := image.GetSnapshot(snapName)

	isProtected, err := snap.IsProtected()
	if err != nil {
		return fmt.Errorf("failed to check if snapshot (%s) is protected for image (%s): %w", snapName, imageName, err)
	}
	if isProtected {
		if err := snap.Unprotect(); err != nil {
			return fmt.Errorf("failed to unprotect snapshot (%s) for image (%s): %w", snapName, imageName, err)
		}
	}

	if err := snap.Remove(); err != nil {
		return fmt.Errorf("failed to remove snapshot (%s) for image (%s): %w", snapName, imageName, err)
	}

	return nil
}

func (rc *RadosConn) RbdSnapList(ctx context.Context, imageSpec ImageSpec) (snaps []SnapInfo, err error) {
	err = rc.Do(ctx, func() error {
		_snaps, err := RbdSnapList(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		snaps = _snaps
		return nil
	})
	return
}

func RbdSnapList(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec) (snapInfos []SnapInfo, err error) {
	namespaceName, poolName, imageName, err := Image(string(imageSpec))
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

	image, err := rbd.OpenImage(ioctx, imageName, rbd.NoSnapshot)
	if err != nil {
		err = fmt.Errorf("failed to open image (%s): %w", imageName, err)
		return
	}
	defer image.Close()

	snaps, err := image.GetSnapshotNames()
	if err != nil {
		err = fmt.Errorf("failed to list snapshots for image (%s): %w", imageName, err)
		return
	}

	snapInfos = make([]SnapInfo, len(snaps))
	for i, snap := range snaps {
		snapshot := image.GetSnapshot(snap.Name)
		protected, protectedErr := snapshot.IsProtected()
		if protectedErr != nil {
			err = fmt.Errorf("failed to check if snapshot (%s) is protected for image (%s): %w", snap.Name, imageName, protectedErr)
			return
		}

		// Best-effort snapshot timestamp.
		// NOTE: ceph tracker #47287 reports older clusters may assert
		// if a non-existing snap ID is supplied. We use IDs returned by
		// GetSnapshotNames() to avoid mismatches and treat retrieval failure
		// as non-fatal by keeping Timestamp as zero value.
		timestamp := time.Time{}
		if snapTs, snapTsErr := image.GetSnapTimestamp(snap.Id); snapTsErr == nil {
			timestamp = time.Unix(snapTs.Sec, snapTs.Nsec)
		}

		snapInfos[i] = SnapInfo{
			ID:        snap.Id,
			Name:      snap.Name,
			Size:      snap.Size,
			Protected: protected,
			Timestamp: timestamp,
		}
	}

	return
}

func RbdSnapInfo(ctx context.Context, conn *rados.Conn, snapSpec SnapSpec) (info *rbd.ImageInfo, err error) {
	namespaceName, poolName, imageName, snapName, err := Snap(string(snapSpec))
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

	image, err := rbd.OpenImage(ioctx, imageName, snapName)
	if err != nil {
		err = fmt.Errorf("failed to open image (%s): %w", imageName, err)
		return
	}
	defer image.Close()

	info, err = image.Stat()
	if err != nil {
		err = fmt.Errorf("failed to stat snapshot (%s) for image (%s): %w", snapName, imageName, err)
		return
	}

	return
}
