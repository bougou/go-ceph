package rados

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	cephrados "github.com/ceph/go-ceph/rados"
)

type RadosConn struct {
	conn    *cephrados.Conn
	mu      sync.RWMutex
	retries int

	cephConfFile string
}

// NewRadosConn creates a RADOS connection wrapper.
//
// cephConfFile is the Ceph config file path; if empty, the default config is used.
// The config is loaded and the connection object is created immediately.
func NewRadosConn(cephConfFile string) (rc *RadosConn, err error) {
	conn, err := newRadosConn(cephConfFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create rados connection: %w", err)
	}

	return &RadosConn{conn: conn, cephConfFile: cephConfFile}, nil
}

func (rc *RadosConn) WithRetries(retries int) *RadosConn {
	rc.retries = retries
	return rc
}

func newRadosConn(cephConfFile string) (conn *cephrados.Conn, err error) {
	conn, err = cephrados.NewConn()
	if err != nil {
		return
	}

	if cephConfFile == "" {
		err = conn.ReadDefaultConfigFile()
		if err != nil {
			return
		}
	} else {
		err = conn.ReadConfigFile(cephConfFile)
		if err != nil {
			return
		}
	}

	return
}

func (rc *RadosConn) Connect() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.conn == nil {
		conn, err := newRadosConn(rc.cephConfFile)
		if err != nil {
			return err
		}
		rc.conn = conn
	}
	return rc.conn.Connect()
}

func (rc *RadosConn) Reconnect() error {
	rc.mu.Lock()
	// Close existing connection if any
	if rc.conn != nil {
		rc.conn.Shutdown()
		rc.conn = nil
	}
	rc.mu.Unlock()

	return rc.Connect()
}

// Test connects to the cluster and verifies the connection is working.
func (rc *RadosConn) Test() error {
	if err := rc.Connect(); err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	rc.mu.RLock()
	defer rc.mu.RUnlock()

	if rc.conn == nil {
		return fmt.Errorf("rados connection is not established")
	}

	if _, err := rc.conn.GetClusterStats(); err != nil {
		return fmt.Errorf("failed to verify cluster connection: %w", err)
	}
	return nil
}

// ensureConnected checks connection health and reconnects if necessary.
func (rc *RadosConn) ensureConnected() error {
	rc.mu.RLock()
	conn := rc.conn
	if conn == nil {
		rc.mu.RUnlock()
		return rc.Reconnect()
	}

	// GetClusterStats is a lightweight operation to verify connection.
	_, err := conn.GetClusterStats()
	rc.mu.RUnlock()
	if err != nil {
		return rc.Reconnect()
	}

	return nil
}

// isConnectionError reports whether err indicates a connection problem.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	// Common connection-related error patterns
	// You may need to adjust based on actual go-ceph error types
	errStr := strings.ToLower(err.Error())
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"no route to host",
		"network is unreachable",
		"timeout",
		"eof",
		"i/o timeout",
	}
	for _, connErr := range connectionErrors {
		if strings.Contains(errStr, connErr) {
			return true
		}
	}
	return false
}

func (rc *RadosConn) Close() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if rc.conn != nil {
		rc.conn.Shutdown()
		rc.conn = nil
	}
	return nil
}

// Do runs operation with automatic reconnection when a connection error occurs.
//
// The connection pointer is held for the entire operation so concurrent Reconnect
// calls cannot nil it out mid-flight.
func (rc *RadosConn) Do(ctx context.Context, operation func(conn *cephrados.Conn) error) error {
	var lastErr error

	maxRetries := rc.retries
	if maxRetries < 0 {
		maxRetries = 0
	}
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Ensure connection is healthy before operation
		if err := rc.ensureConnected(); err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
			continue
		}

		// Execute the operation while holding a read lock so rc.conn stays stable.
		err := func() error {
			rc.mu.RLock()
			defer rc.mu.RUnlock()

			if rc.conn == nil {
				return fmt.Errorf("rados connection is not established")
			}
			return operation(rc.conn)
		}()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is connection-related and should trigger reconnect
		if isConnectionError(err) {
			if reconnErr := rc.Reconnect(); reconnErr != nil {
				lastErr = fmt.Errorf("operation failed: %w, reconnect failed: %v", err, reconnErr)
			}
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		// Non-connection error, return immediately
		return err
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}
