package rbd

import (
	"fmt"
	"strings"
)

const (
	// DefaultPoolName is the default RADOS pool when a spec omits the pool segment.
	DefaultPoolName string = "rbd"
)

// ImageSpec is an RBD image reference in the form "[pool[/namespace]/]image".
// A single segment names the image and uses DefaultPoolName for the pool.
// Valid reports whether the spec has no "@" and non-empty pool and image names.
type ImageSpec string

func NewImageSpec(poolName, imageName string) ImageSpec {
	return ImageSpec(fmt.Sprintf("%s/%s", poolName, imageName))
}

func NewImageSpecWithNamespace(poolName, namespaceName, imageName string) ImageSpec {
	if namespaceName == "" {
		return NewImageSpec(poolName, imageName)
	}
	return ImageSpec(fmt.Sprintf("%s/%s/%s", poolName, namespaceName, imageName))
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

// SnapSpec is an RBD snapshot reference in the form "[pool[/namespace]/]image@snap".
// Valid reports whether the spec contains exactly one "@" and non-empty pool, image, and snap names.
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

// PoolSpec is a pool reference in the form "pool[/namespace]".
type PoolSpec string

func NewPoolSpec(poolName string, namespace string) PoolSpec {
	if namespace == "" {
		return PoolSpec(poolName)
	}
	return PoolSpec(fmt.Sprintf("%s/%s", poolName, namespace))
}

func (p PoolSpec) clean() string {
	s := string(p)
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "/")
	return s
}

func (p PoolSpec) Pool() string {
	parts := strings.Split(p.clean(), "/")
	if len(parts) == 0 || parts[0] == "" {
		return DefaultPoolName
	}
	return parts[0]
}

func (p PoolSpec) Namespace() string {
	parts := strings.Split(p.clean(), "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func (p PoolSpec) Valid() bool {
	s := p.clean()
	if s == "" {
		return false
	}
	if strings.Contains(s, "@") {
		return false
	}
	parts := strings.Split(s, "/")
	if len(parts) > 2 {
		return false
	}
	return p.Pool() != ""
}

func (p PoolSpec) Equal(other PoolSpec) bool {
	return p.clean() == other.clean()
}

// Pool parses a pool spec ("pool[/namespace]") and returns its components.
func Pool(poolSpec string) (pool string, namespace string, err error) {
	spec := strings.TrimSpace(poolSpec)
	if strings.Contains(spec, "@") {
		err = fmt.Errorf("invalid pool spec: %s", poolSpec)
		return
	}
	ps := PoolSpec(spec)
	pool = ps.Pool()
	namespace = ps.Namespace()
	if pool == "" {
		err = fmt.Errorf("invalid pool spec: %s", poolSpec)
	}
	return
}
