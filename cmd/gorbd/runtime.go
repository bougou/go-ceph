package main

import (
	"context"
	"fmt"

	"github.com/bougou/go-ceph/pkg/rados"
)

// withConn runs fn with a RADOS connection.
func withConn(ctx context.Context, fn func(*rados.RadosConn) error) error {
	conn, err := rados.NewRadosConn(globalOpts.cephConf, false)
	if err != nil {
		return fmt.Errorf("failed to create rados connection: %w", err)
	}
	defer conn.Close()
	conn.WithRetries(globalOpts.retries)
	return fn(conn)
}

// withoutConn runs fn without opening a RADOS connection.
func withoutConn(ctx context.Context, fn func() error) error {
	return fn()
}
