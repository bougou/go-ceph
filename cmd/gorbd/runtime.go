package main

import (
	"context"
	"fmt"

	ceph "github.com/bougou/go-ceph"
)

// withConn is a helper function to run a function with a connection.
func withConn(ctx context.Context, fn func(*ceph.RadosConn) error) error {
	conn, err := ceph.NewRadosConn(globalOpts.cephConf, false)
	if err != nil {
		return fmt.Errorf("failed to create rados connection: %w", err)
	}
	defer conn.Close()
	conn.WithRetries(globalOpts.retries)
	return fn(conn)
}

// withoutConn is a helper function to run a function without a connection.
func withoutConn(ctx context.Context, fn func() error) error {
	return fn()
}
