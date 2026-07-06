package rados

import (
	"context"

	"github.com/bougou/go-ceph/pkg/rbd"
	cephrados "github.com/ceph/go-ceph/rados"
)

func (rc *RadosConn) RbdMigrationPrepare(
	ctx context.Context,
	srcImageSpec rbd.ImageSpec,
	dstImageSpec rbd.ImageSpec,
	opts ...rbd.RbdImageOption,
) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdMigrationPrepare(ctx, conn, srcImageSpec, dstImageSpec, opts...)
	})
}

func (rc *RadosConn) RbdMigrationPrepareImport(
	ctx context.Context,
	destImageSpec rbd.ImageSpec,
	sourceSpec string,
	opts ...rbd.RbdImageOption,
) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdMigrationPrepareImport(ctx, conn, destImageSpec, sourceSpec, opts...)
	})
}

func (rc *RadosConn) RbdMigrationExecute(ctx context.Context, imageSpec rbd.ImageSpec, prog *rbd.Progress) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdMigrationExecute(ctx, conn, imageSpec, prog)
	})
}

func (rc *RadosConn) RbdMigrationCommit(ctx context.Context, imageSpec rbd.ImageSpec) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdMigrationCommit(ctx, conn, imageSpec)
	})
}

func (rc *RadosConn) RbdMigrationAbort(ctx context.Context, imageSpec rbd.ImageSpec) error {
	return rc.Do(ctx, func(conn *cephrados.Conn) error {
		return rbd.RbdMigrationAbort(ctx, conn, imageSpec)
	})
}

func (rc *RadosConn) RbdMigrationStatus(ctx context.Context, imageSpec rbd.ImageSpec) (status *rbd.ImageStatusMigration, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		status, err = rbd.RbdMigrationStatus(ctx, conn, imageSpec)
		return err
	})
	return
}
