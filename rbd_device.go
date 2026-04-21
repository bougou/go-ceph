package ceph

import (
	"context"
	"fmt"

	"github.com/bougou/go-ceph/krbd"
	"github.com/ceph/go-ceph/rados"
)

func (rc *RadosConn) RbdDeviceList(ctx context.Context) ([]krbd.Device, error) {
	var devices []krbd.Device = nil
	err := rc.Do(ctx, func() error {
		_devices, err := RbdDeviceList(ctx, rc.conn)
		if err != nil {
			return err
		}
		devices = _devices
		return nil
	})
	return devices, err
}

// RbdDeviceList does not require a connection, so you can pass nil as the connection.
func RbdDeviceList(ctx context.Context, conn *rados.Conn) ([]krbd.Device, error) {
	devices, err := krbd.Devices()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}
	return devices, nil
}

// RbdDeviceFind does not require a connection, so you can pass nil as the connection.
func RbdDeviceFind[T ~string](ctx context.Context, conn *rados.Conn, spec T) (krbd.Device, error) {
	s := string(spec)
	namespace := ""
	pool := ""
	image := ""
	snapshot := ""

	if snapSpec := SnapSpec(s); snapSpec.Valid() {
		namespace = snapSpec.Namespace()
		pool = snapSpec.Pool()
		image = snapSpec.Image()
		snapshot = snapSpec.Snap()
	} else {
		imageSpec := ImageSpec(s)
		if !imageSpec.Valid() {
			return krbd.Device{}, errInvalidImageSpec
		}
		namespace = imageSpec.Namespace()
		pool = imageSpec.Pool()
		image = imageSpec.Image()
	}

	device, err := krbd.Find(namespace, pool, image, snapshot)
	if err != nil {
		return krbd.Device{}, fmt.Errorf("failed to find device: %w", err)
	}
	return device, nil
}

func (rc *RadosConn) RbdDeviceMap(ctx context.Context, imageOrSnapSpec string, options *krbd.Options) error {
	return rc.Do(ctx, func() error {
		return RbdDeviceMap(ctx, rc.conn, imageOrSnapSpec, options)
	})
}

func RbdDeviceMap(ctx context.Context, conn *rados.Conn, imageOrSnapSpec string, options *krbd.Options) error {
	s := string(imageOrSnapSpec)
	namespace := ""
	pool := ""
	image := ""
	snapshot := ""

	if snapSpec := SnapSpec(s); snapSpec.Valid() {
		namespace = snapSpec.Namespace()
		pool = snapSpec.Pool()
		image = snapSpec.Image()
		snapshot = snapSpec.Snap()
	} else {
		imageSpec := ImageSpec(s)
		if !imageSpec.Valid() {
			return errInvalidImageSpec
		}
		namespace = imageSpec.Namespace()
		pool = imageSpec.Pool()
		image = imageSpec.Image()
	}

	monitors, err := getMonHosts(conn)
	if err != nil {
		return fmt.Errorf("getMonHosts failed: %w", err)
	}

	if options == nil {
		options = &krbd.Options{}
	}

	if options.Name == "" {
		options.Name = "admin"
	}

	if options.Secret == "" {
		keyrings, err := getKeyrings(conn)
		if err != nil {
			return fmt.Errorf("getKeyrings failed: %w", err)
		}

		if secret, ok := secretFromKeyringsForAdmin(keyrings); ok {
			options.Secret = secret
		}
	}

	if snapshot != "" {
		options.ReadOnly = true
	}

	mapImage := krbd.Image{
		DevID:     -1,
		Monitors:  monitors,
		Namespace: namespace,
		Pool:      pool,
		Image:     image,
		Snapshot:  snapshot,
		Options:   options,
	}

	mapWriter, err := krbd.MapWriter()
	if err != nil {
		return fmt.Errorf("failed to get default map writer: %w", err)
	}
	defer mapWriter.Close()

	if err := mapImage.Map(mapWriter); err != nil {
		return fmt.Errorf("failed to map image: %w", err)
	}
	return nil
}

func (rc *RadosConn) RbdDeviceUnmap(ctx context.Context, imageOrSnapSpec string, options *krbd.Options) error {
	return rc.Do(ctx, func() error {
		return RbdDeviceUnmap(ctx, rc.conn, imageOrSnapSpec, options)
	})
}

func RbdDeviceUnmap(ctx context.Context, conn *rados.Conn, imageOrSnapSpec string, options *krbd.Options) error {
	s := string(imageOrSnapSpec)
	namespace := ""
	pool := ""
	image := ""
	snapshot := ""

	if snapSpec := SnapSpec(s); snapSpec.Valid() {
		namespace = snapSpec.Namespace()
		pool = snapSpec.Pool()
		image = snapSpec.Image()
		snapshot = snapSpec.Snap()
	} else {
		imageSpec := ImageSpec(s)
		if !imageSpec.Valid() {
			return errInvalidImageSpec
		}
		namespace = imageSpec.Namespace()
		pool = imageSpec.Pool()
		image = imageSpec.Image()
	}

	monitors, err := getMonHosts(conn)
	if err != nil {
		return fmt.Errorf("getMonHosts failed: %w", err)
	}

	if options == nil {
		options = &krbd.Options{}
	}

	if options.Name == "" {
		options.Name = "admin"
	}

	if options.Secret == "" {
		keyrings, err := getKeyrings(conn)
		if err != nil {
			return fmt.Errorf("getKeyrings failed: %w", err)
		}

		if secret, ok := secretFromKeyringsForAdmin(keyrings); ok {
			options.Secret = secret
		}
	}
	mapImage := krbd.Image{
		DevID:     0,
		Monitors:  monitors,
		Namespace: namespace,
		Pool:      pool,
		Image:     image,
		Snapshot:  snapshot,
		Options:   options,
	}

	unmapWriter, err := krbd.UnmapWriter()
	if err != nil {
		return fmt.Errorf("failed to get default unmap writer: %w", err)
	}
	defer unmapWriter.Close()

	if err := mapImage.Unmap(unmapWriter); err != nil {
		return fmt.Errorf("failed to unmap device: %w", err)
	}
	return nil
}

func (rc *RadosConn) RbdDeviceUnmapByID(ctx context.Context, devID int, options *krbd.Options) error {
	return rc.Do(ctx, func() error {
		return RbdDeviceUnmapByID(ctx, rc.conn, devID, options)
	})
}

func RbdDeviceUnmapByID(ctx context.Context, conn *rados.Conn, devID int, options *krbd.Options) error {
	monitors, err := getMonHosts(conn)
	if err != nil {
		return fmt.Errorf("failed to get monitor hosts: %w", err)
	}

	keyrings, err := getKeyrings(conn)
	if err != nil {
		return fmt.Errorf("failed to get keyring data: %w", err)
	}

	if options == nil {
		options = &krbd.Options{}
	}

	if secret, ok := secretFromKeyringsForAdmin(keyrings); ok {
		options.Name = "admin"
		options.Secret = secret
	}

	mapImage := krbd.Image{
		DevID:    devID,
		Monitors: monitors,
		Options:  options,
	}

	unmapWriter, err := krbd.UnmapWriter()
	if err != nil {
		return fmt.Errorf("failed to get default unmap writer: %w", err)
	}
	defer unmapWriter.Close()

	if err := mapImage.Unmap(unmapWriter); err != nil {
		return fmt.Errorf("failed to unmap device: %w", err)
	}
	return nil
}
