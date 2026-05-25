package rbd

import (
	"context"
	"fmt"

	"github.com/bougou/go-ceph/pkg/krbd"
	cephrados "github.com/ceph/go-ceph/rados"
)

// RbdDeviceList does not require a connection, so you can pass nil as the connection.
func RbdDeviceList(ctx context.Context, conn *cephrados.Conn) (devices []krbd.Device, err error) {
	devices, err = krbd.Devices()
	if err != nil {
		err = fmt.Errorf("failed to get devices: %w", err)
		return
	}
	return
}

// RbdDeviceFind does not require a connection, so you can pass nil as the connection.
func RbdDeviceFind[T ~string](ctx context.Context, conn *cephrados.Conn, spec T) (device krbd.Device, err error) {
	namespace, pool, image, snapshot, err := ImageOrSnap(string(spec))
	if err != nil {
		return
	}

	device, err = krbd.Find(namespace, pool, image, snapshot)
	if err != nil {
		err = fmt.Errorf("failed to find device: %w", err)
		return
	}
	return
}

func RbdDeviceMap(ctx context.Context, conn *cephrados.Conn, imageOrSnapSpec string, options *krbd.Options) error {
	namespace, pool, image, snapshot, err := ImageOrSnap(imageOrSnapSpec)
	if err != nil {
		return err
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

func RbdDeviceUnmap(ctx context.Context, conn *cephrados.Conn, imageOrSnapSpec string, options *krbd.Options) error {
	namespace, pool, image, snapshot, err := ImageOrSnap(imageOrSnapSpec)
	if err != nil {
		return err
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

func RbdDeviceUnmapByID(ctx context.Context, conn *cephrados.Conn, devID int, options *krbd.Options) error {
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
