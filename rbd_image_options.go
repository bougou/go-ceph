package ceph

import (
	"fmt"

	"github.com/ceph/go-ceph/rbd"
)

// RbdImageOptionFn configures a *rbd.ImageOptions before create / clone / copy.
// Functions run in order; later entries override earlier ones for the same option
// keys. Unset keys use librbd / cluster defaults (same as passing an empty
// rbd.ImageOptions to go-ceph).
//
// For any key not covered by RbdOpt* helpers, use RbdOptUint64 / RbdOptString with
// github.com/ceph/go-ceph/rbd.ImageOption constants, or pass a custom func that calls
// (*rbd.ImageOptions).SetUint64 / SetString.
//
// Feature bits: use RbdOptFeatures for the full mask, RbdOptDefault for
// rbd.RbdFeaturesDefault plus format 2, RbdOptFeaturesClear for
// RBD_IMAGE_OPTION_FEATURES_CLEAR, or RbdOptFeaturesSet for FEATURES_SET.
type RbdImageOptionFn func(*rbd.ImageOptions) error

func rbdImageOptionsFromFns(fns ...RbdImageOptionFn) (opts *rbd.ImageOptions, err error) {
	opts = rbd.NewRbdImageOptions()

	for _, fn := range fns {
		if fn == nil {
			continue
		}
		callErr := fn(opts)
		if callErr != nil {
			opts.Destroy()
			err = callErr
			return
		}
	}
	return
}

// RbdOptUint64 sets an arbitrary librbd image option by key (rbd.ImageOption) and
// uint64 value. Use this with official go-ceph constants, e.g. rbd.ImageOptionFormat.
func RbdOptUint64(option rbd.ImageOption, value uint64) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetUint64(option, value)
	}
}

// RbdOptString sets an arbitrary librbd image option by key and string value
// (e.g. rbd.ImageOptionDataPool, rbd.ImageOptionJournalPool).
func RbdOptString(option rbd.ImageOption, value string) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetString(option, value)
	}
}

// RbdOptFeatures sets RBD_IMAGE_OPTION_FEATURES (replaces any value set earlier
// in the same option list by a previous RbdOptFeatures or RbdOptUint64 on that key).
func RbdOptFeatures(features uint64) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetUint64(rbd.ImageOptionFeatures, features)
	}
}

// RbdOptOrder sets RBD_IMAGE_OPTION_ORDER (object size = 2^order bytes).
func RbdOptOrder(order uint64) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetUint64(rbd.ImageOptionOrder, order)
	}
}

// RbdOptFormat sets RBD_IMAGE_OPTION_FORMAT (e.g. 1 or 2).
func RbdOptFormat(format uint64) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetUint64(rbd.ImageOptionFormat, format)
	}
}

// RbdOptDefault sets image format 2 and RBD_IMAGE_OPTION_FEATURES to
// rbd.RbdFeaturesDefault (the C macro RBD_FEATURES_DEFAULT from the librbd
// version you link against at build/runtime). This does not read ceph.conf
// rbd_default_features; for that behavior omit options or use cluster tooling.
func RbdOptDefault() RbdImageOptionFn {
	return RbdOptCompose(
		RbdOptFormat(2),
		RbdOptFeatures(rbd.RbdFeaturesDefault),
	)
}

// RbdOptStripeUnit sets RBD_IMAGE_OPTION_STRIPE_UNIT.
func RbdOptStripeUnit(unit uint64) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetUint64(rbd.ImageOptionStripeUnit, unit)
	}
}

// RbdOptStripeCount sets RBD_IMAGE_OPTION_STRIPE_COUNT.
func RbdOptStripeCount(count uint64) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetUint64(rbd.ImageOptionStripeCount, count)
	}
}

// RbdOptDataPool sets RBD_IMAGE_OPTION_DATA_POOL.
func RbdOptDataPool(pool string) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetString(rbd.ImageOptionDataPool, pool)
	}
}

// RbdOptCloneFormat sets RBD_IMAGE_OPTION_CLONE_FORMAT (clone / copy clone step).
func RbdOptCloneFormat(format uint64) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetUint64(rbd.ImageOptionCloneFormat, format)
	}
}

// RbdOptFeaturesSet sets RBD_IMAGE_OPTION_FEATURES_SET.
func RbdOptFeaturesSet(mask uint64) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetUint64(rbd.ImageOptionFeaturesSet, mask)
	}
}

// RbdOptFeaturesClear sets RBD_IMAGE_OPTION_FEATURES_CLEAR. Librbd clears the
// bits in mask from the effective features together with RBD_IMAGE_OPTION_FEATURES
// (and FEATURES_SET).
func RbdOptFeaturesClear(mask uint64) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetUint64(rbd.ImageOptionFeaturesClear, mask)
	}
}

// RbdOptJournalOrder sets RBD_IMAGE_OPTION_JOURNAL_ORDER.
func RbdOptJournalOrder(order uint64) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetUint64(rbd.ImageOptionJournalOrder, order)
	}
}

// RbdOptJournalSplayWidth sets RBD_IMAGE_OPTION_JOURNAL_SPLAY_WIDTH.
func RbdOptJournalSplayWidth(width uint64) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetUint64(rbd.ImageOptionJournalSplayWidth, width)
	}
}

// RbdOptJournalPool sets RBD_IMAGE_OPTION_JOURNAL_POOL.
func RbdOptJournalPool(pool string) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetString(rbd.ImageOptionJournalPool, pool)
	}
}

// RbdOptFlatten sets RBD_IMAGE_OPTION_FLATTEN.
func RbdOptFlatten(v uint64) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		return o.SetUint64(rbd.ImageOptionFlatten, v)
	}
}

// RbdOptCompose runs several option functions in order.
func RbdOptCompose(fns ...RbdImageOptionFn) RbdImageOptionFn {
	return func(o *rbd.ImageOptions) error {
		for _, fn := range fns {
			if fn == nil {
				continue
			}
			if err := fn(o); err != nil {
				return fmt.Errorf("RbdOptCompose: %w", err)
			}
		}
		return nil
	}
}
