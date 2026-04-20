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
