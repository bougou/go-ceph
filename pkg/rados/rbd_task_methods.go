package rados

import (
	"context"

	"github.com/bougou/go-ceph/pkg/rbd"
	cephrados "github.com/ceph/go-ceph/rados"
)

func (rc *RadosConn) RbdTaskAddFlatten(ctx context.Context, imageSpec rbd.ImageSpec) (task rbd.TaskResponse, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		task, err = rbd.RbdTaskAddFlatten(ctx, conn, imageSpec)
		return err
	})
	return
}

func (rc *RadosConn) RbdTaskAddRemove(ctx context.Context, imageSpec rbd.ImageSpec) (task rbd.TaskResponse, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		task, err = rbd.RbdTaskAddRemove(ctx, conn, imageSpec)
		return err
	})
	return
}

func (rc *RadosConn) RbdTaskAddTrashRemove(ctx context.Context, imageIDSpec rbd.ImageSpec) (task rbd.TaskResponse, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		task, err = rbd.RbdTaskAddTrashRemove(ctx, conn, imageIDSpec)
		return err
	})
	return
}

func (rc *RadosConn) RbdTaskAddMigrationExecute(ctx context.Context, imageSpec rbd.ImageSpec) (task rbd.TaskResponse, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		task, err = rbd.RbdTaskAddMigrationExecute(ctx, conn, imageSpec)
		return err
	})
	return
}

func (rc *RadosConn) RbdTaskAddMigrationCommit(ctx context.Context, imageSpec rbd.ImageSpec) (task rbd.TaskResponse, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		task, err = rbd.RbdTaskAddMigrationCommit(ctx, conn, imageSpec)
		return err
	})
	return
}

func (rc *RadosConn) RbdTaskAddMigrationAbort(ctx context.Context, imageSpec rbd.ImageSpec) (task rbd.TaskResponse, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		task, err = rbd.RbdTaskAddMigrationAbort(ctx, conn, imageSpec)
		return err
	})
	return
}

func (rc *RadosConn) RbdTaskList(ctx context.Context) (tasks []rbd.TaskResponse, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		tasks, err = rbd.RbdTaskList(ctx, conn)
		return err
	})
	return
}

func (rc *RadosConn) RbdTaskGet(ctx context.Context, taskID string) (task rbd.TaskResponse, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		task, err = rbd.RbdTaskGet(ctx, conn, taskID)
		return err
	})
	return
}

func (rc *RadosConn) RbdTaskCancel(ctx context.Context, taskID string) (task rbd.TaskResponse, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		task, err = rbd.RbdTaskCancel(ctx, conn, taskID)
		return err
	})
	return
}
