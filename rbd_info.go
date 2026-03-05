package ceph

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

// RbdInfo retrieves detailed information about an RBD image.
// If the image does not exist, it returns nil, nil.
func (rc *RadosConn) RbdInfo(ctx context.Context, imageSpec ImageSpec) (*RbdImageInfo, error) {
	var info *RbdImageInfo = nil

	err := rc.Do(ctx, func() error {
		_info, err := RbdInfo(ctx, rc.conn, imageSpec)
		if err != nil {
			return err
		}
		info = _info
		return nil
	})

	return info, err
}

// RbdInfo retrieves detailed information about an RBD image.
// If the image does not exist, it returns nil, nil.
func RbdInfo(ctx context.Context, conn *rados.Conn, imageSpec ImageSpec) (*RbdImageInfo, error) {
	if !imageSpec.Valid() {
		return nil, errInvalidImageSpec
	}

	poolName := imageSpec.Pool()
	imageName := imageSpec.Image()
	namespaceName := imageSpec.Namespace()

	// Open IOContext for the pool
	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	image, err := rbd.OpenImage(ioctx, imageName, rbd.NoSnapshot)
	if err != nil {
		if isErrNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	return ConvertRbdImageToRbdInfo(image)
}

// RbdImageInfo contains detailed information about an RBD image,
// equivalent to the output of `rbd info <image>`.
type RbdImageInfo struct {
	// Name is the name of the RBD image
	Name string `json:"name,omitempty"`

	// Size is the size of the image in bytes
	Size uint64 `json:"size,omitempty"`

	// NumObjects is the number of objects the image is divided into
	NumObjects uint64 `json:"num_objects,omitempty"`

	// Order is the object order (power of 2), determines object size (2^order bytes)
	Order int `json:"order,omitempty"`

	// ObjectSize is the size of each object in bytes (derived from order: 2^order)
	ObjectSize uint64 `json:"object_size,omitempty"`

	// SnapshotCount is the number of snapshots for this image
	SnapshotCount int `json:"snapshot_count,omitempty"`

	// ID is the internal image ID
	ID string `json:"id,omitempty"`

	// BlockNamePrefix is the prefix for the underlying RADOS objects
	BlockNamePrefix string `json:"block_name_prefix,omitempty"`

	// Format is the RBD image format version (1 or 2)
	Format int `json:"format,omitempty"`

	// Features is the bitmask of enabled features
	Features uint64 `json:"features,omitempty"`

	// FeatureNames is the list of enabled feature names
	FeatureNames []string `json:"feature_names,omitempty"`

	// CreateTimestamp is when the image was created
	CreateTimestamp time.Time `json:"create_timestamp,omitempty		"`

	// AccessTimestamp is when the image was last accessed
	AccessTimestamp time.Time `json:"access_timestamp,omitempty"`

	// ModifyTimestamp is when the image was last modified
	ModifyTimestamp time.Time `json:"modify_timestamp,omitempty"`

	// Parent is the parent snapshot name
	Parent string `json:"parent,omitempty,omitzero"`

	// Overlap is the number of bytes that are shared between the image and its parent
	Overlap uint64 `json:"overlap,omitempty,omitzero"`
}

func sizeHuman(size uint64, precision int) string {
	const (
		KiB = 1024
		MiB = KiB * 1024
		GiB = MiB * 1024
		TiB = GiB * 1024
	)

	var value float64
	var unit string = "B"
	switch {
	case size >= TiB:
		value = float64(size) / float64(TiB)
		unit = "TiB"
	case size >= GiB:
		value = float64(size) / float64(GiB)
		unit = "GiB"
	case size >= MiB:
		value = float64(size) / float64(MiB)
		unit = "MiB"
	case size >= KiB:
		value = float64(size) / float64(KiB)
		unit = "KiB"
	default:
		value = float64(size)
		unit = "B"
	}

	if precision <= 0 {
		// Round up to the nearest integer
		return fmt.Sprintf("%d %s", int(math.Ceil(value)), unit)
	}

	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, value) + " " + unit
}

// ObjectSizeHuman returns a human-readable object size string (e.g., "4 MiB")
func (r *RbdImageInfo) ObjectSizeHuman() string {
	const (
		KiB = 1024
		MiB = KiB * 1024
	)

	switch {
	case r.ObjectSize >= MiB:
		return fmt.Sprintf("%d MiB", r.ObjectSize/MiB)
	case r.ObjectSize >= KiB:
		return fmt.Sprintf("%d KiB", r.ObjectSize/KiB)
	default:
		return fmt.Sprintf("%d B", r.ObjectSize)
	}
}

// String returns a formatted string representation similar to `rbd info` output
func (r *RbdImageInfo) String() string {
	out := fmt.Sprintf(`rbd image '%s':
	size %s in %d objects
	order %d (%s objects)
	snapshot_count: %d
	id: %s
	block_name_prefix: %s
	format: %d
	features: %v
	create_timestamp: %s
	access_timestamp: %s
	modify_timestamp: %s`,
		r.Name,
		sizeHuman(r.Size, 0), r.NumObjects,
		r.Order, r.ObjectSizeHuman(),
		r.SnapshotCount,
		r.ID,
		r.BlockNamePrefix,
		r.Format,
		strings.Join(r.FeatureNames, ", "),
		r.CreateTimestamp.Format(time.ANSIC),
		r.AccessTimestamp.Format(time.ANSIC),
		r.ModifyTimestamp.Format(time.ANSIC),
	)

	if r.Parent != "" {
		out += fmt.Sprintf("\nparent: %s", r.Parent)
	}
	if r.Overlap != 0 {
		out += fmt.Sprintf("\noverlap: %s", sizeHuman(r.Overlap, 0))
	}
	return out

}

// ConvertRbdImageToRbdInfo retrieves detailed information from an already opened RBD image.
func ConvertRbdImageToRbdInfo(image *rbd.Image) (*RbdImageInfo, error) {
	imageName := image.GetName()
	info := &RbdImageInfo{
		Name: imageName,
	}

	// Get basic image stats
	stat, err := image.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat image (%s): %w", imageName, err)
	}
	info.Size = stat.Size
	info.NumObjects = stat.Num_objs
	info.Order = stat.Order
	info.ObjectSize = stat.Obj_size
	info.BlockNamePrefix = stat.Block_name_prefix

	// Get image ID
	id, err := image.GetId()
	if err != nil {
		return nil, fmt.Errorf("failed to get image ID for image (%s): %w", imageName, err)
	}
	info.ID = id

	// Get image format (old format = 1, new format = 2)
	isOldFormat, err := image.IsOldFormat()
	if err != nil {
		return nil, fmt.Errorf("failed to get image format for image (%s): %w", imageName, err)
	}
	if isOldFormat {
		info.Format = 1
	} else {
		info.Format = 2
	}

	// Get features
	features, err := image.GetFeatures()
	if err != nil {
		return nil, fmt.Errorf("failed to get image features for image (%s): %w", imageName, err)
	}
	info.Features = features

	// Convert features bitmask to feature names
	featureSet := rbd.FeatureSet(features)
	info.FeatureNames = featureSet.Names()

	// Get snapshot count
	snapshots, err := image.GetSnapshotNames()
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot names for image (%s): %w", imageName, err)
	}
	info.SnapshotCount = len(snapshots)

	// Get timestamps
	createTs, err := image.GetCreateTimestamp()
	if err != nil {
		return nil, fmt.Errorf("failed to get create timestamp for image (%s): %w", imageName, err)
	}
	info.CreateTimestamp = time.Unix(createTs.Sec, createTs.Nsec)

	accessTs, err := image.GetAccessTimestamp()
	if err != nil {
		return nil, fmt.Errorf("failed to get access timestamp for image (%s): %w", imageName, err)
	}
	info.AccessTimestamp = time.Unix(accessTs.Sec, accessTs.Nsec)

	modifyTs, err := image.GetModifyTimestamp()
	if err != nil {
		return nil, fmt.Errorf("failed to get modify timestamp for image (%s): %w", imageName, err)
	}
	info.ModifyTimestamp = time.Unix(modifyTs.Sec, modifyTs.Nsec)

	parent, err := image.GetParent()
	if err != nil {
		if isErrNotFound(err) {
			return info, nil
		}
		return nil, fmt.Errorf("failed to get parent for image (%s): %w", imageName, err)
	} else {
		info.Parent = fmt.Sprintf("%s/%s@%s", parent.Image.PoolName, parent.Image.ImageName, parent.Snap.SnapName)

		overlap, err := image.GetOverlap()
		if err != nil {
			return nil, fmt.Errorf("failed to get overlap for image (%s): %w", imageName, err)
		}
		info.Overlap = overlap
	}

	return info, nil
}
