package rbd

import (
	"context"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

// PersistentCacheStateKey is the image metadata key used by librbd for
// persistent write-back cache state (see rbd status "Persistent cache state").
const PersistentCacheStateKey = ".rbd_persistent_cache_state"

// RbdImageMetaGet returns the metadata value for key.
// Equivalent to: rbd image-meta get <image-spec> <key>
func RbdImageMetaGet(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec, key string) (string, error) {
	var value string
	err := withOpenImage(conn, imageSpec, true, func(image *cephrbd.Image) error {
		v, err := image.GetMetadata(key)
		if err != nil {
			return fmt.Errorf("failed to get metadata key %q: %w", key, err)
		}
		value = v
		return nil
	})
	return value, err
}

// RbdImageMetaList returns all metadata key/value pairs for the image.
// Equivalent to: rbd image-meta list <image-spec>
func RbdImageMetaList(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec) (map[string]string, error) {
	var meta map[string]string
	err := withOpenImage(conn, imageSpec, true, func(image *cephrbd.Image) error {
		m, err := image.ListMetadata()
		if err != nil {
			return fmt.Errorf("failed to list metadata: %w", err)
		}
		meta = m
		return nil
	})
	return meta, err
}

// RbdImageMetaSet sets metadata key to value.
// Equivalent to: rbd image-meta set <image-spec> <key> <value>
func RbdImageMetaSet(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec, key, value string) error {
	return withOpenImage(conn, imageSpec, false, func(image *cephrbd.Image) error {
		if err := image.SetMetadata(key, value); err != nil {
			return fmt.Errorf("failed to set metadata key %q: %w", key, err)
		}
		return nil
	})
}

// RbdImageMetaRemove removes the metadata key.
// Equivalent to: rbd image-meta remove <image-spec> <key>
func RbdImageMetaRemove(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec, key string) error {
	return withOpenImage(conn, imageSpec, false, func(image *cephrbd.Image) error {
		if err := image.RemoveMetadata(key); err != nil {
			return fmt.Errorf("failed to remove metadata key %q: %w", key, err)
		}
		return nil
	})
}

func withOpenImage(
	conn *cephrados.Conn,
	imageSpec ImageSpec,
	readOnly bool,
	fn func(image *cephrbd.Image) error,
) error {
	namespaceName, poolName, imageName, err := Image(string(imageSpec))
	if err != nil {
		return err
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()
	ioctx.SetNamespace(namespaceName)

	var image *cephrbd.Image
	if readOnly {
		image, err = cephrbd.OpenImageReadOnly(ioctx, imageName, cephrbd.NoSnapshot)
	} else {
		image, err = cephrbd.OpenImage(ioctx, imageName, cephrbd.NoSnapshot)
	}
	if err != nil {
		return fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	return fn(image)
}
