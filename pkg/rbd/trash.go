package rbd

import (
	"context"
	"fmt"
	"os"
	"time"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

type Trash struct {
	ID             string              `json:"id" xml:"id"`
	Name           string              `json:"name" xml:"name"`
	DeletedAt      time.Time           `json:"deleted_at,omitempty" xml:"deleted_at,omitempty"`
	DefermentEndAt time.Time           `json:"deferment_end_at,omitempty" xml:"deferment_end_at,omitempty"`
	Status         string              `json:"status,omitempty" xml:"status,omitempty"`
	Parent         *cephrbd.ParentInfo `json:"parent,omitempty" xml:"parent,omitempty"`
}

func trashStatus(defermentEndTime time.Time) string {
	now := time.Now()
	timeStr := defermentEndTime.Local().Format(time.ANSIC)
	if now.Before(defermentEndTime) {
		return "protected until " + timeStr
	}
	return "expired at " + timeStr
}

func trashParent(ioctx *cephrados.IOContext, imageID string) *cephrbd.ParentInfo {
	image, err := cephrbd.OpenImageByIdReadOnly(ioctx, imageID, cephrbd.NoSnapshot)
	if err != nil {
		if !isErrNotFound(err) {
			fmt.Fprintf(os.Stderr, "rbd: error opening %s: %v\n", imageID, err)
		}
		return nil
	}

	parentInfo, err := image.GetParent()
	_ = image.Close()
	if err != nil {
		if !isErrNotFound(err) {
			fmt.Fprintf(os.Stderr, "rbd: error opening %s: %v\n", imageID, err)
		}
		return nil
	}

	return parentInfo
}

// RbdTrashList returns trash images in the given pool/namespace.
func RbdTrashList(ctx context.Context, conn *cephrados.Conn, poolSpec PoolSpec) ([]Trash, error) {
	poolName, namespaceName, err := Pool(string(poolSpec))
	if err != nil {
		return nil, err
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	entries, err := cephrbd.GetTrashList(ioctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list trash images in pool (%s): %w", poolName, err)
	}

	out := make([]Trash, 0, len(entries))
	for _, entry := range entries {
		item := Trash{
			ID:             entry.Id,
			Name:           entry.Name,
			DeletedAt:      entry.DeletionTime,
			DefermentEndAt: entry.DefermentEndTime,
			Status:         trashStatus(entry.DefermentEndTime),
			Parent:         trashParent(ioctx, entry.Id),
		}
		out = append(out, item)
	}

	return out, nil
}
