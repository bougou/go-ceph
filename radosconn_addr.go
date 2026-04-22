package ceph

import (
	"fmt"
	"strconv"
	"strings"
)

type address struct {
	addrType addrType
	host     string
	port     int32
	nonce    int
}

type addrType string

const (
	addrTypeV1  addrType = "v1"
	addrTypeV2  addrType = "v2"
	addrTypeAny addrType = "TYPE_ANY"
)

func (a address) String() string {
	b, err := a.MarshalText()
	if err != nil {
		return ""
	}
	return string(b)
}

func (a address) MarshalText() (text []byte, err error) {
	var b strings.Builder
	switch a.addrType {
	case addrTypeV1:
		b.WriteString("v1:")
	case addrTypeV2:
		b.WriteString("v2:")
	}
	b.WriteString(a.host)
	if a.port > 0 {
		b.WriteString(fmt.Sprintf(":%d", a.port))
	}
	if a.nonce > 0 {
		b.WriteString(fmt.Sprintf("/%d", a.nonce))
	}
	text = []byte(b.String())
	return
}

func MarshalAddress(a address) string {
	return a.String()
}

func MarshalAddresses(addrs []address) string {
	var b strings.Builder
	for _, a := range addrs {
		b.WriteString(a.String())
		b.WriteString(",")
	}
	return b.String()
}

func MarshalAddresses2(addrs [][]address) string {
	var b strings.Builder
	for _, group := range addrs {
		b.WriteString("[")
		for _, a := range group {
			b.WriteString(a.String())
			b.WriteString(",")
		}
		b.WriteString("]")
	}
	return b.String()
}

// monAddrs examples:
//   - Three monitor groups (bracketed): [v2:10.97.145.7:3300,v1:10.97.145.7:6789],[v2:10.97.167.34:3300,v1:10.97.167.34:6789],[v2:10.97.166.34:3300,v1:10.97.166.34:6789]
//   - One group without outer brackets: v2:10.97.145.7:3300,v1:10.97.145.7:6789
//   - Flat list across monitors (no brackets): v2:a:3300,v1:a:6789,v2:b:3300,v1:b:6789 — same host endpoints are merged into one inner slice, in first-seen host order.
//
// Return value: outer slice = one logical monitor per element; inner slice = parsed endpoints for that monitor (e.g. v1 and v2 to the same IP).
//
// see: https://docs.ceph.com/en/nautilus/rados/configuration/msgr2/#address-formats
func parseAddresses(addrs string) (out [][]address, err error) {
	s := strings.TrimSpace(addrs)
	if s == "" {
		return
	}

	for _, top := range splitTopLevelMonGroups(s) {
		top = strings.TrimSpace(top)
		if top == "" {
			continue
		}
		tokens := splitMonAddrList(top)
		if len(tokens) == 0 {
			continue
		}
		grouped, groupErr := groupMonTokensByHost(tokens)
		if groupErr != nil {
			err = groupErr
			return
		}
		out = append(out, grouped...)
	}
	if len(out) == 0 {
		return
	}
	return
}

// formatMonitorAddr renders a parsed address in mon_host / krbd style.

// splitTopLevelMonGroups splits mon_host into bracket-separated monitor groups.
// Commas inside "[...]" (e.g. IPv6) are not group separators.
func splitTopLevelMonGroups(s string) []string {
	if strings.Contains(s, "],[") {
		parts := strings.Split(s, "],[")
		for i := range parts {
			p := strings.TrimSpace(parts[i])
			if i == 0 {
				p = strings.TrimPrefix(p, "[")
			}
			if i == len(parts)-1 {
				p = strings.TrimSuffix(p, "]")
			}
			parts[i] = p
		}
		return parts
	}
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") && strings.Contains(s, ",") {
		return []string{strings.TrimSpace(s[1 : len(s)-1])}
	}
	return []string{s}
}

// splitMonAddrList splits a single group's address list on commas, respecting nested brackets (IPv6).
func splitMonAddrList(s string) []string {
	var out []string
	var b strings.Builder
	depth := 0
	for _, r := range s {
		switch r {
		case '[':
			depth++
			b.WriteRune(r)
		case ']':
			if depth > 0 {
				depth--
			}
			b.WriteRune(r)
		case ',':
			if depth == 0 {
				if tok := strings.TrimSpace(b.String()); tok != "" {
					out = append(out, tok)
				}
				b.Reset()
			} else {
				b.WriteRune(r)
			}
		default:
			b.WriteRune(r)
		}
	}
	if tok := strings.TrimSpace(b.String()); tok != "" {
		out = append(out, tok)
	}
	return out
}

// groupMonTokensByHost merges tokens that share the same host; host order follows first occurrence in tokens.
func groupMonTokensByHost(tokens []string) (out [][]address, err error) {
	order := make([]string, 0)
	byHost := make(map[string][]address)
	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		a, parseErr := parseAddress(tok)
		if parseErr != nil {
			err = fmt.Errorf("parseAddress (%q): %w", tok, parseErr)
			return
		}
		host := a.host
		if _, ok := byHost[host]; !ok {
			order = append(order, host)
		}
		byHost[host] = append(byHost[host], a)
	}
	out = make([][]address, 0, len(order))
	for _, h := range order {
		out = append(out, byHost[h])
	}
	return
}

// addr can be:
//   - full path format [type:]host[:port]/[nonce]
//   - e.g. like v2:10.0.0.10:3300/0
//   - the 'type', 'port', 'nonce' are optional.
func (a *address) UnmarshalText(text []byte) error {
	parsed := address{
		addrType: addrTypeAny,
	}

	addr := string(text)
	s := strings.TrimSpace(addr)
	if s == "" {
		return fmt.Errorf("empty monitor address")
	}

	// Optional nonce: [type:]host[:port]/[nonce]
	if i := strings.LastIndex(s, "/"); i >= 0 {
		if i == len(s)-1 {
			return fmt.Errorf("invalid monitor nonce in %q", addr)
		}
		nonce, err := strconv.Atoi(s[i+1:])
		if err != nil || nonce < 0 {
			return fmt.Errorf("invalid monitor nonce in %q", addr)
		}
		parsed.nonce = nonce
		s = s[:i]
	}

	if strings.HasPrefix(s, string(addrTypeV1)+":") {
		parsed.addrType = addrTypeV1
		s = s[len(addrTypeV1)+1:]
	} else if strings.HasPrefix(s, string(addrTypeV2)+":") {
		parsed.addrType = addrTypeV2
		s = s[len(addrTypeV2)+1:]
	}

	if s == "" {
		return fmt.Errorf("missing monitor host in %q", addr)
	}

	parsePort := func(portStr string) (int32, error) {
		port, err := strconv.Atoi(portStr)
		if err != nil || port <= 0 || port > 65535 {
			return 0, fmt.Errorf("invalid monitor port in %q", addr)
		}
		return int32(port), nil
	}

	if strings.HasPrefix(s, "[") {
		// Bracketed host (typically IPv6): [host] or [host]:port
		end := strings.Index(s, "]")
		if end <= 1 {
			return fmt.Errorf("invalid bracketed monitor host in %q", addr)
		}
		parsed.host = s[1:end]
		rest := s[end+1:]
		if rest == "" {
			*a = parsed
			return nil
		}
		if !strings.HasPrefix(rest, ":") || len(rest) == 1 {
			return fmt.Errorf("invalid monitor address %q", addr)
		}
		port, err := parsePort(rest[1:])
		if err != nil {
			return err
		}
		parsed.port = port
		*a = parsed
		return nil
	}

	if strings.Count(s, ":") == 1 {
		host, portStr, ok := strings.Cut(s, ":")
		if !ok || host == "" || portStr == "" {
			return fmt.Errorf("invalid monitor address %q", addr)
		}
		port, err := parsePort(portStr)
		if err != nil {
			return err
		}
		parsed.host = host
		parsed.port = port
		*a = parsed
		return nil
	}

	// Host only, including unbracketed IPv6 without port.
	parsed.host = s
	if parsed.host == "" {
		return fmt.Errorf("missing monitor host in %q", addr)
	}
	*a = parsed
	return nil
}

func parseAddress(addr string) (a address, err error) {
	err = a.UnmarshalText([]byte(addr))
	if err != nil {
		return
	}
	return
}
