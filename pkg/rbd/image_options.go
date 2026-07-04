package rbd

import (
	"fmt"

	cephrbd "github.com/ceph/go-ceph/rbd"
)

// RbdImageOption configures a *cephrbd.ImageOptions before create, clone, or copy.
// Functions run in order; later entries override earlier ones for the same option
// keys. Unset keys use librbd or cluster defaults (same as passing an empty
// cephrbd.ImageOptions to go-ceph).
//
// For any key not covered by RbdOpt* helpers, use RbdOptUint64 or RbdOptString with
// github.com/ceph/go-ceph/rbd.ImageOption constants, or pass a custom func that calls
// (*cephrbd.ImageOptions).SetUint64 or SetString.
//
// Feature bits: use RbdOptFeatures for the full mask, RbdOptDefault for
// cephrbd.RbdFeaturesDefault plus format 2, RbdOptFeaturesClear for
// RBD_IMAGE_OPTION_FEATURES_CLEAR, or RbdOptFeaturesSet for FEATURES_SET.
type RbdImageOption func(*cephrbd.ImageOptions) error

func rbdImageOptions(options ...RbdImageOption) (opts *cephrbd.ImageOptions, err error) {
	opts = cephrbd.NewRbdImageOptions()

	for _, fn := range options {
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

// RbdOptUint64 sets an arbitrary librbd image option by key (cephrbd.ImageOption) and
// uint64 value. Use this with official go-ceph constants, e.g. cephrbd.ImageOptionFormat.
func RbdOptUint64(option cephrbd.ImageOption, value uint64) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetUint64(option, value)
	}
}

// RbdOptString sets an arbitrary librbd image option by key and string value
// (e.g. cephrbd.ImageOptionDataPool, cephrbd.ImageOptionJournalPool).
func RbdOptString(option cephrbd.ImageOption, value string) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetString(option, value)
	}
}

// RbdOptFeatures sets RBD_IMAGE_OPTION_FEATURES (replaces any value set earlier
// in the same option list by a previous RbdOptFeatures or RbdOptUint64 on that key).
func RbdOptFeatures(features uint64) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetUint64(cephrbd.ImageOptionFeatures, features)
	}
}

// RbdOptOrder sets RBD_IMAGE_OPTION_ORDER (object size = 2^order bytes).
func RbdOptOrder(order uint64) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetUint64(cephrbd.ImageOptionOrder, order)
	}
}

// RbdOptFormat sets RBD_IMAGE_OPTION_FORMAT (e.g. 1 or 2).
func RbdOptFormat(format uint64) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetUint64(cephrbd.ImageOptionFormat, format)
	}
}

// RbdOptDefault sets image format 2 and RBD_IMAGE_OPTION_FEATURES to
// cephrbd.RbdFeaturesDefault (the C macro RBD_FEATURES_DEFAULT from the librbd
// version you link against at build/runtime). This does not read ceph.conf
// rbd_default_features; for that behavior omit options or use cluster tooling.
func RbdOptDefault() RbdImageOption {
	return RbdOptCompose(
		RbdOptFormat(2),
		RbdOptFeatures(cephrbd.RbdFeaturesDefault),
	)
}

// RbdOptStripeUnit sets RBD_IMAGE_OPTION_STRIPE_UNIT.
func RbdOptStripeUnit(unit uint64) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetUint64(cephrbd.ImageOptionStripeUnit, unit)
	}
}

// RbdOptStripeCount sets RBD_IMAGE_OPTION_STRIPE_COUNT.
func RbdOptStripeCount(count uint64) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetUint64(cephrbd.ImageOptionStripeCount, count)
	}
}

// RbdOptDataPool sets RBD_IMAGE_OPTION_DATA_POOL.
func RbdOptDataPool(pool string) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetString(cephrbd.ImageOptionDataPool, pool)
	}
}

// RbdOptCloneFormat sets RBD_IMAGE_OPTION_CLONE_FORMAT (clone or copy-clone step).
func RbdOptCloneFormat(format uint64) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetUint64(cephrbd.ImageOptionCloneFormat, format)
	}
}

// RbdOptFeaturesSet sets RBD_IMAGE_OPTION_FEATURES_SET.
func RbdOptFeaturesSet(mask uint64) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetUint64(cephrbd.ImageOptionFeaturesSet, mask)
	}
}

// RbdOptFeaturesClear sets RBD_IMAGE_OPTION_FEATURES_CLEAR. Librbd clears the
// bits in mask from the effective features together with RBD_IMAGE_OPTION_FEATURES
// (and FEATURES_SET).
func RbdOptFeaturesClear(mask uint64) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetUint64(cephrbd.ImageOptionFeaturesClear, mask)
	}
}

// RbdOptJournalOrder sets RBD_IMAGE_OPTION_JOURNAL_ORDER.
func RbdOptJournalOrder(order uint64) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetUint64(cephrbd.ImageOptionJournalOrder, order)
	}
}

// RbdOptJournalSplayWidth sets RBD_IMAGE_OPTION_JOURNAL_SPLAY_WIDTH.
func RbdOptJournalSplayWidth(width uint64) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetUint64(cephrbd.ImageOptionJournalSplayWidth, width)
	}
}

// RbdOptJournalPool sets RBD_IMAGE_OPTION_JOURNAL_POOL.
func RbdOptJournalPool(pool string) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetString(cephrbd.ImageOptionJournalPool, pool)
	}
}

// RbdOptFlatten sets RBD_IMAGE_OPTION_FLATTEN.
func RbdOptFlatten(v uint64) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
		return o.SetUint64(cephrbd.ImageOptionFlatten, v)
	}
}

// RbdOptCompose runs several option functions in order.
func RbdOptCompose(fns ...RbdImageOption) RbdImageOption {
	return func(o *cephrbd.ImageOptions) error {
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
