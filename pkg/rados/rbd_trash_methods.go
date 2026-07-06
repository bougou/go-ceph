package rados

import (
	"context"

	"github.com/bougou/go-ceph/pkg/rbd"
	cephrados "github.com/ceph/go-ceph/rados"
)

func (rc *RadosConn) RbdTrashList(ctx context.Context, poolSpec rbd.PoolSpec) (entries []rbd.Trash, err error) {
	err = rc.Do(ctx, func(conn *cephrados.Conn) error {
		_entries, err := rbd.RbdTrashList(ctx, conn, poolSpec)
		if err != nil {
			return err
		}
		entries = _entries
		return nil
	})
	return
}
