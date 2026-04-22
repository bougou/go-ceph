package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func parseSizeToBytes(input string) (sizeBytes uint64, err error) {
	s := strings.TrimSpace(strings.ToUpper(input))
	if s == "" {
		err = errors.New("size is empty")
		return
	}

	multiplier := uint64(1)
	switch {
	case strings.HasSuffix(s, "KIB"):
		multiplier = 1 << 10
		s = strings.TrimSuffix(s, "KIB")
	case strings.HasSuffix(s, "MIB"):
		multiplier = 1 << 20
		s = strings.TrimSuffix(s, "MIB")
	case strings.HasSuffix(s, "GIB"):
		multiplier = 1 << 30
		s = strings.TrimSuffix(s, "GIB")
	case strings.HasSuffix(s, "TIB"):
		multiplier = 1 << 40
		s = strings.TrimSuffix(s, "TIB")
	case strings.HasSuffix(s, "KB"), strings.HasSuffix(s, "K"):
		multiplier = 1 << 10
		s = strings.TrimSuffix(strings.TrimSuffix(s, "KB"), "K")
	case strings.HasSuffix(s, "MB"), strings.HasSuffix(s, "M"):
		multiplier = 1 << 20
		s = strings.TrimSuffix(strings.TrimSuffix(s, "MB"), "M")
	case strings.HasSuffix(s, "GB"), strings.HasSuffix(s, "G"):
		multiplier = 1 << 30
		s = strings.TrimSuffix(strings.TrimSuffix(s, "GB"), "G")
	case strings.HasSuffix(s, "TB"), strings.HasSuffix(s, "T"):
		multiplier = 1 << 40
		s = strings.TrimSuffix(strings.TrimSuffix(s, "TB"), "T")
	case strings.HasSuffix(s, "B"):
		s = strings.TrimSuffix(s, "B")
	}

	s = strings.TrimSpace(s)
	if s == "" {
		err = fmt.Errorf("invalid size %q", input)
		return
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		err = fmt.Errorf("invalid size %q: %w", input, err)
		return
	}
	sizeBytes = n * multiplier
	return
}
