package rbd

import (
	"fmt"
	"strconv"
	"strings"
)

// rbdMaxParentChainLen is Ceph's RBD_MAX_PARENT_CHAIN_LEN: the clone-on-write
// parent chain is capped at this many levels. Cloning beyond this depth fails.
const rbdMaxParentChainLen = 16

// rbdMaxParentDepth is the maximum value returned by len(RbdParents): a child
// image can sit at most rbdMaxParentChainLen-1 levels below the original base.
const rbdMaxParentDepth = rbdMaxParentChainLen - 1

type copyConfig struct {
	imageOptions []RbdImageOption
}

// CopyOption configures RbdCopy and RbdCopySnap.
type CopyOption func(*copyConfig)

func copyConfigFrom(options ...CopyOption) copyConfig {
	cfg := copyConfig{}
	for _, opt := range options {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// CopyWithImageOptions sets librbd image options for the destination clone.
func CopyWithImageOptions(opts ...RbdImageOption) CopyOption {
	return func(c *copyConfig) {
		c.imageOptions = append(c.imageOptions, opts...)
	}
}

type cloneConfig struct {
	imageOptions     []RbdImageOption
	autoFlattenDepth *uint8 // nil = disabled; flatten when parent depth > threshold
}

// CloneOption configures RbdClone.
type CloneOption func(*cloneConfig)

func cloneConfigFrom(options ...CloneOption) (cloneConfig, error) {
	cfg := cloneConfig{}
	for _, opt := range options {
		if opt != nil {
			opt(&cfg)
		}
	}
	if err := cfg.validate(); err != nil {
		return cloneConfig{}, err
	}
	return cfg, nil
}

func (c cloneConfig) validate() error {
	if c.autoFlattenDepth == nil {
		return nil
	}
	if int(*c.autoFlattenDepth) > rbdMaxParentDepth {
		return fmt.Errorf("parent depth threshold must be at most %d (RBD parent chain limit)", rbdMaxParentDepth)
	}
	return nil
}

// CloneWithImageOptions sets librbd image options for the destination clone.
func CloneWithImageOptions(opts ...RbdImageOption) CloneOption {
	return func(c *cloneConfig) {
		c.imageOptions = append(c.imageOptions, opts...)
	}
}

// CloneWithoutAutoFlatten disables auto-flatten after clone. This is the default
// when no auto-flatten option is passed; use explicitly to override a prior option.
func CloneWithoutAutoFlatten() CloneOption {
	return func(c *cloneConfig) {
		c.autoFlattenDepth = nil
	}
}

// CloneWithAutoFlattenDepth submits a background flatten task when parent depth
// is greater than n. n=0 always flattens (a fresh clone always has depth at least 1).
// n must be between 0 and rbdMaxParentDepth (15) inclusive.
func CloneWithAutoFlattenDepth(n uint8) CloneOption {
	return func(c *cloneConfig) {
		threshold := n
		c.autoFlattenDepth = &threshold
	}
}

// FlattenTask identifies a background flatten task submitted by RbdClone when
// auto-flatten is triggered. Returned only when a task is submitted; otherwise
// the API returns nil.
type FlattenTask struct {
	ID string `json:"id,omitempty"`
}

// ParseCloneAutoFlattenOption parses a clone auto-flatten mode string into a
// CloneOption for CLI use.
//
// Supported values: "", "none", or a decimal depth threshold 0..15.
// Empty string returns nil (no option). "none" returns CloneWithoutAutoFlatten.
// A number N enables auto-flatten when parent depth > N.
func ParseCloneAutoFlattenOption(s string) (CloneOption, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if s == "none" {
		return CloneWithoutAutoFlatten(), nil
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, fmt.Errorf("invalid clone flatten mode %q: use none or a depth threshold 0..%d", s, rbdMaxParentDepth)
	}
	if n < 0 || n > rbdMaxParentDepth {
		return nil, fmt.Errorf("parent depth threshold must be between 0 and %d (RBD parent chain limit)", rbdMaxParentDepth)
	}

	return CloneWithAutoFlattenDepth(uint8(n)), nil
}
