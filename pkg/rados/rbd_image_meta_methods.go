package rados

import (
	"context"

	"github.com/bougou/go-ceph/pkg/rbd"
	cephrados "github.com/ceph/go-ceph/rados"
)

func (rc *RadosConn) RbdImageMetaGet(ctx context.Context, imageSpec rbd.ImageSpec, key string) (value string, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		value, err = rbd.RbdImageMetaGet(ctx, conn, imageSpec, key)
		return err
	})
	return
}

func (rc *RadosConn) RbdImageMetaList(ctx context.Context, imageSpec rbd.ImageSpec) (meta map[string]string, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		meta, err = rbd.RbdImageMetaList(ctx, conn, imageSpec)
		return err
	})
	return
}

func (rc *RadosConn) RbdImageMetaSet(ctx context.Context, imageSpec rbd.ImageSpec, key, value string) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdImageMetaSet(ctx, conn, imageSpec, key, value)
	})
}

func (rc *RadosConn) RbdImageMetaRemove(ctx context.Context, imageSpec rbd.ImageSpec, key string) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdImageMetaRemove(ctx, conn, imageSpec, key)
	})
}
