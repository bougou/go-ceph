package ceph

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
)

// secretFromKeyringsForAdmin returns the key for the default RADOS user (client.admin).
// Ceph keyring sections are usually named "client.admin"; some configs use "admin".
func secretFromKeyringsForAdmin(keyrings map[string]string) (secret string, ok bool) {
	if s, ok := keyrings["client.admin"]; ok && s != "" {
		return s, true
	}
	s, ok := keyrings["admin"]
	if ok && s != "" {
		return s, true
	}
	return "", false
}

func expandKeyringPaths(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == ':'
	})
	out := make([]string, 0, len(parts)+3)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	out = append(out,
		"/etc/ceph/ceph.client.admin.keyring",
		"/etc/ceph/ceph.keyring",
		"/etc/ceph/keyring",
	)
	return lo.Uniq(out)
}

func parseCephKeyring(path string) (out map[string]string, err error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return
	}
	defer f.Close()

	out = map[string]string{}
	scanner := bufio.NewScanner(f)
	entity := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			entity = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		if entity == "" {
			continue
		}
		if strings.HasPrefix(line, "key") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				out[entity] = strings.TrimSpace(parts[1])
			}
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		err = scanErr
		return
	}
	return
}
