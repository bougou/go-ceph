package main

import (
	"context"
	"fmt"

	ceph "github.com/bougou/go-ceph"
)

func withConn(ctx context.Context, fn func(*ceph.RadosConn) error) error {
	conn, err := ceph.NewRadosConn(globalOpts.cephConf, false)
	if err != nil {
		return fmt.Errorf("failed to create rados connection: %w", err)
	}
	defer conn.Close()
	conn.WithRetries(globalOpts.retries)
	return fn(conn)
}
