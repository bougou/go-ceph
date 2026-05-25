package rbd

import (
	"context"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

// ImageListEntry is one entry in the long-form image list output.
// One entry per image, plus one additional entry per snapshot of that image.
type ImageListEntry struct {
	Image     string `json:"image" xml:"image"`
	Snapshot  string `json:"snapshot,omitempty" xml:"snapshot,omitempty"`
	Size      uint64 `json:"size" xml:"size"`
	Format    int    `json:"format" xml:"format"`
	Parent    string `json:"parent,omitempty" xml:"parent,omitempty"`
	Protected string `json:"protected,omitempty" xml:"protected,omitempty"`
}

// RbdList returns the names of all RBD images in the given pool/namespace.
func RbdList(ctx context.Context, conn *cephrados.Conn, poolSpec PoolSpec) (images []string, err error) {
	poolName, namespaceName, err := Pool(string(poolSpec))
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

	images, err = cephrbd.GetImageNames(ioctx)
	if err != nil {
		err = fmt.Errorf("failed to list images in pool (%s): %w", poolName, err)
		return
	}
	return
}

// RbdListLong returns one entry per image (and one extra entry per snapshot)
// with size, format, parent, and protection info, equivalent to "rbd ls -l".
func RbdListLong(ctx context.Context, conn *cephrados.Conn, poolSpec PoolSpec) (entries []ImageListEntry, err error) {
	poolName, namespaceName, err := Pool(string(poolSpec))
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

	names, err := cephrbd.GetImageNames(ioctx)
	if err != nil {
		err = fmt.Errorf("failed to list images in pool (%s): %w", poolName, err)
		return
	}

	for _, name := range names {
		imageEntries, listErr := listOneImage(ioctx, name)
		if listErr != nil {
			// Skip images that disappeared between listing and opening
			// (race with concurrent rbd rm / trash move). Matches `rbd ls -l`.
			if isErrNotFound(listErr) {
				continue
			}
			err = listErr
			return
		}
		entries = append(entries, imageEntries...)
	}
	return
}

func listOneImage(ioctx *cephrados.IOContext, name string) (entries []ImageListEntry, err error) {
	image, err := cephrbd.OpenImageReadOnly(ioctx, name, cephrbd.NoSnapshot)
	if err != nil {
		err = fmt.Errorf("failed to open image (%s): %w", name, err)
		return
	}
	defer image.Close()

	stat, err := image.Stat()
	if err != nil {
		err = fmt.Errorf("failed to stat image (%s): %w", name, err)
		return
	}

	format := 2
	if isOld, oldErr := image.IsOldFormat(); oldErr == nil && isOld {
		format = 1
	}

	parent := ""
	if parentInfo, parentErr := image.GetParent(); parentErr == nil {
		parent = fmt.Sprintf("%s/%s@%s", parentInfo.Image.PoolName, parentInfo.Image.ImageName, parentInfo.Snap.SnapName)
	}

	entries = append(entries, ImageListEntry{
		Image:  name,
		Size:   stat.Size,
		Format: format,
		Parent: parent,
	})

	snaps, snapErr := image.GetSnapshotNames()
	if snapErr != nil {
		err = fmt.Errorf("failed to list snapshots for image (%s): %w", name, snapErr)
		return
	}

	for _, snap := range snaps {
		snapshot := image.GetSnapshot(snap.Name)
		protected := ""
		if isProtected, protErr := snapshot.IsProtected(); protErr == nil {
			if isProtected {
				protected = "true"
			} else {
				protected = "false"
			}
		}
		entries = append(entries, ImageListEntry{
			Image:     name,
			Snapshot:  snap.Name,
			Size:      snap.Size,
			Format:    format,
			Parent:    parent,
			Protected: protected,
		})
	}
	return
}
