package rbd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cephrados "github.com/ceph/go-ceph/rados"
	"github.com/samber/lo"
)

func getMonAddrs(conn *cephrados.Conn) (monAddrs [][]address, err error) {
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
func getMonHosts(conn *cephrados.Conn) (out []string, err error) {
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

func getKeyrings(conn *cephrados.Conn) (keyrings map[string]string, err error) {
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
