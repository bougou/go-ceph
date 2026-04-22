package ceph

import (
	"errors"
	"fmt"
	"strings"
)

const (
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

func (i ImageSpec) Valid() bool {
	s := i.clean()
	// image spec must not include snapshot delimiter.
	if strings.Contains(s, "@") {
		return false
	}
	return i.Image() != "" && i.Pool() != ""
}

func (i ImageSpec) Equal(other ImageSpec) bool {
	return i.clean() == other.clean()
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

func (v SnapSpec) Valid() bool {
	s := v.clean()
	// snap spec must include exactly one snapshot delimiter.
	if strings.Count(s, "@") != 1 {
		return false
	}
	return v.Snap() != "" && v.Image() != "" && v.Pool() != ""
}

func (v SnapSpec) Equal(other SnapSpec) bool {
	return v.clean() == other.clean()
}

// ImageOrSnap parses an image or snapshot spec and returns the namespace, pool, image, and snapshot.
// If the returned snapshot is empty, it means the spec is an image spec.
func ImageOrSnap(imageOrSnapSpec string) (namespace string, pool string, image string, snapshot string, err error) {
	s := strings.TrimSpace(imageOrSnapSpec)
	if strings.Contains(s, "@") {
		namespace, pool, image, snapshot, err = Snap(s)
		if err != nil {
			err = fmt.Errorf("invalid image or snapshot spec: %s", s)
		}
		return
	}
	namespace, pool, image, err = Image(s)
	if err != nil {
		err = fmt.Errorf("invalid image or snapshot spec: %s", s)
	}
	return
}

func Image(imageSpec string) (namespace string, pool string, image string, err error) {
	spec := strings.TrimSpace(imageSpec)
	if strings.Contains(spec, "@") {
		err = fmt.Errorf("invalid image spec: %s", imageSpec)
		return
	}
	imageSpecValue := ImageSpec(spec)
	namespace = imageSpecValue.Namespace()
	pool = imageSpecValue.Pool()
	image = imageSpecValue.Image()
	if image == "" || pool == "" {
		err = fmt.Errorf("invalid image spec: %s", imageSpec)
	}
	return
}

func Snap(snapSpec string) (namespace string, pool string, image string, snapshot string, err error) {
	spec := strings.TrimSpace(snapSpec)
	if strings.Count(spec, "@") != 1 {
		err = fmt.Errorf("invalid snap spec: %s", snapSpec)
		return
	}
	snapSpecValue := SnapSpec(spec)
	namespace = snapSpecValue.Namespace()
	pool = snapSpecValue.Pool()
	image = snapSpecValue.Image()
	snapshot = snapSpecValue.Snap()
	if snapshot == "" || image == "" || pool == "" {
		err = fmt.Errorf("invalid snap spec: %s", snapSpec)
	}
	return
}
