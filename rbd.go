package ceph

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ceph/go-ceph/rbd"
)

const (
	// Default image order is 22 (2^22 = 4 MiB objects)
	DefaultImageOrder uint64 = 22

	// Default image features is layering and deep flatten
	DefaultImageFeatures uint64 = uint64(rbd.FeatureLayering) | uint64(rbd.FeatureDeepFlatten)

	// Default pool name
	DefaultPoolName string = "rbd"
)

// ErrPlatformNotSupported is returned when RBD operations are called on unsupported platforms
var ErrPlatformNotSupported = errors.New("RBD is not supported on this platform")

type ImageSpec string

func NewImageSpec(poolName string, imageName string) ImageSpec {
	return ImageSpec(fmt.Sprintf("%s/%s", poolName, imageName))
}

func (i ImageSpec) clean() string {
	s := string(i)
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "/")
	return s
}

func (i ImageSpec) Pool() string {
	parts := strings.Split(i.clean(), "/")
	if len(parts) == 0 || len(parts) == 1 {
		return DefaultPoolName
	}
	return parts[0]
}

func (i ImageSpec) Namespace() string {
	s := i.clean()
	parts := strings.Split(s, "/")
	if len(parts) <= 2 {
		return ""
	}
	return parts[1]
}

func (i ImageSpec) Image() string {
	s := i.clean()
	parts := strings.Split(s, "/")
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	if len(parts) == 2 {
		return parts[1]
	}
	return parts[2]
}

type SnapSpec string

func NewSnapSpec(poolName string, imageName string, snapName string) SnapSpec {
	return SnapSpec(fmt.Sprintf("%s/%s@%s", poolName, imageName, snapName))
}

func (snapSepc SnapSpec) clean() string {
	s := string(snapSepc)
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "/")
	return s
}

func (v SnapSpec) Snap() string {
	s := v.clean()
	parts := strings.Split(s, "@")
	if len(parts) == 0 || len(parts) == 1 {
		return ""
	}
	return parts[1]
}

func (v SnapSpec) Pool() string {
	s := v.clean()
	parts := strings.Split(s, "@")
	if len(parts) == 0 {
		return ""
	}

	imageSpec := ImageSpec(parts[0])
	return imageSpec.Pool()
}

func (v SnapSpec) Image() string {
	s := v.clean()
	parts := strings.Split(s, "@")
	if len(parts) == 0 {
		return ""
	}

	imageSpec := ImageSpec(parts[0])
	return imageSpec.Image()
}

func (v SnapSpec) Namespace() string {
	s := v.clean()
	parts := strings.Split(s, "@")
	if len(parts) == 0 {
		return ""
	}

	imageSpec := ImageSpec(parts[0])
	return imageSpec.Namespace()
}

// RbdImageInfo contains detailed information about an RBD image,
// equivalent to the output of `rbd info <image>`.
type RbdImageInfo struct {
	// Name is the name of the RBD image
	Name string `json:"name"`

	// Size is the size of the image in bytes
	Size uint64 `json:"size"`

	// NumObjects is the number of objects the image is divided into
	NumObjects uint64 `json:"num_objects"`

	// Order is the object order (power of 2), determines object size (2^order bytes)
	Order int `json:"order"`

	// ObjectSize is the size of each object in bytes (derived from order: 2^order)
	ObjectSize uint64 `json:"object_size"`

	// SnapshotCount is the number of snapshots for this image
	SnapshotCount int `json:"snapshot_count"`

	// ID is the internal image ID
	ID string `json:"id"`

	// BlockNamePrefix is the prefix for the underlying RADOS objects
	BlockNamePrefix string `json:"block_name_prefix"`

	// Format is the RBD image format version (1 or 2)
	Format int `json:"format"`

	// Features is the bitmask of enabled features
	Features uint64 `json:"features"`

	// FeatureNames is the list of enabled feature names
	FeatureNames []string `json:"feature_names"`

	// CreateTimestamp is when the image was created
	CreateTimestamp time.Time `json:"create_timestamp"`

	// AccessTimestamp is when the image was last accessed
	AccessTimestamp time.Time `json:"access_timestamp"`

	// ModifyTimestamp is when the image was last modified
	ModifyTimestamp time.Time `json:"modify_timestamp"`
}

// SizeHuman returns a human-readable size string (e.g., "100 GiB")
func (r *RbdImageInfo) SizeHuman(precision int) string {
	const (
		KiB = 1024
		MiB = KiB * 1024
		GiB = MiB * 1024
		TiB = GiB * 1024
	)

	var value float64
	var unit string = "B"
	switch {
	case r.Size >= TiB:
		value = float64(r.Size) / float64(TiB)
		unit = "TiB"
	case r.Size >= GiB:
		value = float64(r.Size) / float64(GiB)
		unit = "GiB"
	case r.Size >= MiB:
		value = float64(r.Size) / float64(MiB)
		unit = "MiB"
	case r.Size >= KiB:
		value = float64(r.Size) / float64(KiB)
		unit = "KiB"
	default:
		value = float64(r.Size)
		unit = "B"
	}

	if precision <= 0 {
		return fmt.Sprintf("%d %s", int(value), unit)
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
	return fmt.Sprintf(`rbd image '%s':
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
		r.SizeHuman(0), r.NumObjects,
		r.Order, r.ObjectSizeHuman(),
		r.SnapshotCount,
		r.ID,
		r.BlockNamePrefix,
		r.Format,
		r.FeatureNames,
		r.CreateTimestamp.Format(time.ANSIC),
		r.AccessTimestamp.Format(time.ANSIC),
		r.ModifyTimestamp.Format(time.ANSIC),
	)
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

	return info, nil
}
