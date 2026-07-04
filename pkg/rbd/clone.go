package rbd

import (
	"context"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

// RbdClone clones a snapshot to a new image in the same pool and namespace.
//
// Use CloneWithAutoFlattenDepth to submit a background flatten task when parent
// depth exceeds a threshold (depth > n; n=0 always flattens). CloneWithoutAutoFlatten
// disables auto-flatten (the default when no
// option is passed). The call returns immediately after clone (and task submission
// if triggered); it does not wait for flatten or clean up snapshots.
// Returns a non-nil FlattenTask when a flatten task is submitted.
func RbdClone(ctx context.Context, conn *cephrados.Conn, srcSnapSpec SnapSpec, dstImageSpec ImageSpec, opts ...CloneOption) (*FlattenTask, error) {
	cfg, err := cloneConfigFrom(opts...)
	if err != nil {
		return nil, err
	}

	srcNamespaceName, srcPoolName, srcImageName, srcSnapName, err := Snap(string(srcSnapSpec))
	if err != nil {
		return nil, err
	}

	dstNamespaceName, dstPoolName, dstImageName, err := Image(string(dstImageSpec))
	if err != nil {
		return nil, err
	}

	if srcPoolName != dstPoolName {
		return nil, fmt.Errorf("source pool (%s) and destination pool (%s) are different", srcPoolName, dstPoolName)
	}

	if srcNamespaceName != dstNamespaceName {
		return nil, fmt.Errorf("source namespace (%s) and destination namespace (%s) are different", srcNamespaceName, dstNamespaceName)
	}

	srcIOCtx, err := conn.OpenIOContext(srcPoolName)
	if err != nil {
		return nil, fmt.Errorf("failed to open source pool (%s): %w", srcPoolName, err)
	}
	defer srcIOCtx.Destroy()

	srcIOCtx.SetNamespace(srcNamespaceName)

	srcSnap, err := cephrbd.OpenImage(srcIOCtx, srcImageName, srcSnapName)
	if err != nil {
		return nil, fmt.Errorf("failed to open source snapshot (%s) for image (%s): %w", srcSnapName, srcImageName, err)
	}
	defer srcSnap.Close()

	dstIOCtx, err := conn.OpenIOContext(dstPoolName)
	if err != nil {
		return nil, fmt.Errorf("failed to open destination pool (%s): %w", dstPoolName, err)
	}
	defer dstIOCtx.Destroy()

	dstIOCtx.SetNamespace(dstNamespaceName)

	imageOpts, err := rbdImageOptions(cfg.imageOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to build image options: %w", err)
	}
	defer imageOpts.Destroy()

	if err := cephrbd.CloneImage(srcIOCtx, srcImageName, srcSnapName, dstIOCtx, dstImageName, imageOpts); err != nil {
		return nil, fmt.Errorf("failed to clone image (%s) from snapshot (%s): %w", dstImageName, srcSnapSpec, err)
	}

	if cfg.autoFlattenDepth == nil {
		return nil, nil
	}

	parents, err := RbdParents(ctx, conn, dstImageSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to list parents for image (%s): %w", dstImageSpec, err)
	}
	if uint8(len(parents)) <= *cfg.autoFlattenDepth {
		return nil, nil
	}

	task, err := RbdTaskAddFlatten(ctx, conn, dstImageSpec)
	if err != nil {
		return nil, err
	}
	return &FlattenTask{ID: task.ID}, nil
}
