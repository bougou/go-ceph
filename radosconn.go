package ceph

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ceph/go-ceph/rados"
	"github.com/samber/lo"
)

type RadosConn struct {
	conn    *rados.Conn
	mu      sync.RWMutex
	retries int

	cephConfFile string
	monAddrs     [][]address       // per-monitor parsed endpoints (e.g. v1+v2 for same host grouped)
	keyrings     map[string]string // entity -> secret key
}

// NewRadosConn creates a RADOS connection wrapper.
//
// The `cephConfFile` specifies the Ceph config file path; if empty, the default config is used.
//
// The `lazy` controls whether the underlying connection is created immediately.
//   - When `lazy` is false, the the config is loaded and the connection is created immediately.
//   - When `lazy` is true, config loading and connection creation are deferred to the first `Connect`/`Do` call.
//
// Error behavior:
//   - When `lazy` is false, may return an error immediately if config loading fails.
//   - When `lazy` is true, always returns nil error here
func NewRadosConn(cephConfFile string, lazy bool) (rc *RadosConn, err error) {
	var conn *rados.Conn = nil

	if !lazy {
		var newConn *rados.Conn
		newConn, err = newRadosConn(cephConfFile)
		if err != nil {
			err = fmt.Errorf("failed to create rados connection: %w", err)
			return
		}
		conn = newConn
	}

	rc = &RadosConn{conn: conn, cephConfFile: cephConfFile}
	return
}

func (rc *RadosConn) WithRetries(retries int) *RadosConn {
	rc.retries = retries
	return rc
}

func newRadosConn(cephConfFile string) (conn *rados.Conn, err error) {
	conn, err = rados.NewConn()
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
	if rc.conn == nil {
		conn, err := newRadosConn(rc.cephConfFile)
		if err != nil {
			return err
		}
		rc.conn = conn
	}
	if err := rc.conn.Connect(); err != nil {
		return err
	}
	if err := rc.loadMetadata(); err != nil {
		return fmt.Errorf("loadMetadata failed: %w", err)
	}
	return nil
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

func (rc *RadosConn) loadMetadata() error {
	monAddrs, err := getMonAddrs(rc.conn)
	if err != nil {
		return fmt.Errorf("failed to parse mon_host: %w", err)
	}

	keyrings, err := getKeyrings(rc.conn)
	if err != nil {
		return fmt.Errorf("failed to get keyring data: %w", err)
	}

	rc.mu.Lock()
	rc.monAddrs = monAddrs
	rc.keyrings = keyrings
	rc.mu.Unlock()

	return nil
}

func getMonAddrs(conn *rados.Conn) (monAddrs [][]address, err error) {
	rawMonAddrs, err := conn.GetConfigOption("mon_host")
	if err != nil {
		err = fmt.Errorf("failed to get mon_host: %w", err)
		return
	}

	// The monAddrs is like:
	// [v2:10.97.145.7:3300,v1:10.97.145.7:6789],[v2:10.97.167.34:3300,v1:10.97.167.34:6789],[v2:10.97.166.34:3300,v1:10.97.166.34:6789]

	monAddrs, err = parseAddresses(rawMonAddrs)
	return
}

// getMonHosts returns ONLY the hostnames part of the monitors.
func getMonHosts(conn *rados.Conn) (out []string, err error) {
	groups, err := getMonAddrs(conn)
	if err != nil {
		return
	}
	for _, g := range groups {
		for _, a := range g {
			out = append(out, a.host)
		}
	}
	out = lo.Uniq(out)
	return
}

func getKeyrings(conn *rados.Conn) (keyrings map[string]string, err error) {
	keyringPath, _ := conn.GetConfigOption("keyring")
	paths := expandKeyringPaths(keyringPath)
	keyrings = map[string]string{}
	for _, path := range paths {
		clean := filepath.Clean(path)
		data, parseErr := parseCephKeyring(clean)
		if parseErr != nil {
			if errors.Is(parseErr, os.ErrNotExist) {
				continue
			}
			err = fmt.Errorf("read keyring %s: %w", clean, parseErr)
			return
		}
		for entity, secret := range data {
			keyrings[entity] = secret
		}
	}
	if len(keyrings) == 0 {
		err = fmt.Errorf("no keyring data found after trying: %s", strings.Join(paths, ", "))
		return
	}
	return
}
