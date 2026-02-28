package ceph

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ceph/go-ceph/rados"
)

type RadosConn struct {
	conn    *rados.Conn
	mu      sync.RWMutex
	retries int

	cephConfFile string
}

func NewRadosConn(cephConfFile string, lazy bool) (*RadosConn, error) {
	var conn *rados.Conn = nil

	if !lazy {
		_conn, err := newRadosConn(cephConfFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create rados connection: %w", err)
		}
		conn = _conn
	}

	return &RadosConn{conn: conn, cephConfFile: cephConfFile}, nil
}

func (rc *RadosConn) WithRetries(retries int) *RadosConn {
	rc.retries = retries
	return rc
}

func newRadosConn(cephConfFile string) (*rados.Conn, error) {
	conn, err := rados.NewConn()
	if err != nil {
		return nil, err
	}

	if cephConfFile == "" {
		if err := conn.ReadDefaultConfigFile(); err != nil {
			return nil, err
		}
	} else {
		if err := conn.ReadConfigFile(cephConfFile); err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func (rc *RadosConn) Connect() error {
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
	defer rc.mu.Unlock()

	// Close existing connection if any
	if rc.conn != nil {
		rc.conn.Shutdown()
		rc.conn = nil
	}

	return rc.Connect()
}

// ensureConnected checks connection health and reconnects if necessary
func (rc *RadosConn) ensureConnected() error {
	rc.mu.RLock()
	conn := rc.conn
	rc.mu.RUnlock()

	if conn == nil {
		return rc.Reconnect()
	}

	// Try a simple operation to check if connection is alive
	// GetClusterStats is a lightweight operation to verify connection
	_, err := conn.GetClusterStats()
	if err != nil {
		return rc.Reconnect()
	}

	return nil
}

func (rc *RadosConn) getConn() *rados.Conn {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.conn
}

// isConnectionError checks if the error is related to connection issues
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

// DoWithRetry executes an operation with automatic reconnection on failure
func (rc *RadosConn) Do(ctx context.Context, operation func() error) error {
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

		// Execute the operation
		err := operation()
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
