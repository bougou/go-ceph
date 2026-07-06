package rbd

import (
	"context"
	"fmt"
	"strings"
	"time"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

// RbdMigrationPrepare prepares a live migration from source to destination.
// Equivalent to: rbd migration prepare <source-image-spec> [<dest-image-spec>]
//
// When dstImageSpec is empty or invalid, the destination reuses the source
// pool, namespace, and image name (on-disk layout change with same name).
func RbdMigrationPrepare(
	ctx context.Context,
	conn *cephrados.Conn,
	srcImageSpec ImageSpec,
	dstImageSpec ImageSpec,
	opts ...RbdImageOption,
) error {
	if !srcImageSpec.Valid() {
		return fmt.Errorf("invalid source image spec: %s", srcImageSpec)
	}

	dst := dstImageSpec
	if !dst.Valid() {
		dst = srcImageSpec
	}

	srcNamespace, srcPool, srcImage, err := Image(string(srcImageSpec))
	if err != nil {
		return err
	}
	dstNamespace, dstPool, dstImage, err := Image(string(dst))
	if err != nil {
		return err
	}

	srcIOCtx, err := conn.OpenIOContext(srcPool)
	if err != nil {
		return fmt.Errorf("failed to open source pool (%s): %w", srcPool, err)
	}
	defer srcIOCtx.Destroy()
	srcIOCtx.SetNamespace(srcNamespace)

	dstIOCtx, err := conn.OpenIOContext(dstPool)
	if err != nil {
		return fmt.Errorf("failed to open destination pool (%s): %w", dstPool, err)
	}
	defer dstIOCtx.Destroy()
	dstIOCtx.SetNamespace(dstNamespace)

	imageOpts, err := rbdImageOptions(opts...)
	if err != nil {
		return fmt.Errorf("failed to build image options: %w", err)
	}
	defer imageOpts.Destroy()

	if err := validateSourceForMigrationPrepare(conn, srcIOCtx, srcImage, srcImageSpec); err != nil {
		return err
	}

	if err := cephrbd.MigrationPrepare(srcIOCtx, srcImage, dstIOCtx, dstImage, imageOpts); err != nil {
		if isErrExist(err) {
			return migrationPrepareExistError(conn, dstPool, dstNamespace, dstImage, dst)
		}
		return fmt.Errorf("failed to prepare migration (%s -> %s): %w", srcImageSpec, dst, err)
	}
	return nil
}

func migrationPrepareExistError(
	conn *cephrados.Conn,
	dstPool, dstNamespace, dstImage string,
	dst ImageSpec,
) error {
	ioctx, err := conn.OpenIOContext(dstPool)
	if err != nil {
		return fmt.Errorf("failed to prepare migration to %s: image name %q already exists", dst, dstImage)
	}
	defer ioctx.Destroy()
	ioctx.SetNamespace(dstNamespace)

	entries, err := cephrbd.GetTrashList(ioctx)
	if err == nil {
		for _, entry := range entries {
			if entry.Name == dstImage {
				return fmt.Errorf(
					"failed to prepare migration to %s: image name %q is reserved by trashed image %s (remove from trash or choose another name)",
					dst, dstImage, entry.Id,
				)
			}
		}
	}

	exist, err := RbdExist(context.Background(), conn, dst)
	if err == nil && exist {
		return fmt.Errorf("failed to prepare migration to %s: destination image already exists", dst)
	}

	return fmt.Errorf(
		"failed to prepare migration to %s: image name %q already exists",
		dst, dstImage,
	)
}

func validateSourceForMigrationPrepare(
	conn *cephrados.Conn,
	ioctx *cephrados.IOContext,
	imageName string,
	imageSpec ImageSpec,
) error {
	image, err := cephrbd.OpenImageReadOnly(ioctx, imageName, cephrbd.NoSnapshot)
	if err != nil {
		return fmt.Errorf("failed to open source image (%s): %w", imageSpec, err)
	}
	defer image.Close()

	features, err := image.GetFeatures()
	if err != nil {
		return fmt.Errorf("failed to get source image features (%s): %w", imageSpec, err)
	}
	if features&cephrbd.FeatureMigrating == 0 {
		return nil
	}

	migration, err := readImageMigrationStatus(conn, ioctx, imageName, image)
	if err != nil {
		return fmt.Errorf("source image %s is involved in migration but status is unavailable: %w", imageSpec, err)
	}
	if migration == nil {
		return fmt.Errorf("source image %s has migrating feature but no migration status", imageSpec)
	}

	return fmt.Errorf(
		"source image %s is already involved in migration (source: %s, destination: %s, state: %s); commit or abort it before preparing a new migration",
		imageSpec,
		migration.Source,
		migration.Destination,
		MigrationStateLine(migration),
	)
}

// RbdMigrationPrepareImport prepares an import-only live migration.
// Equivalent to: rbd migration prepare --import-only --source-spec <json> <dest-image-spec>
func RbdMigrationPrepareImport(
	ctx context.Context,
	conn *cephrados.Conn,
	destImageSpec ImageSpec,
	sourceSpec string,
	opts ...RbdImageOption,
) error {
	if !destImageSpec.Valid() {
		return fmt.Errorf("invalid destination image spec: %s", destImageSpec)
	}
	if sourceSpec == "" {
		return fmt.Errorf("source spec is required for import-only migration")
	}

	dstNamespace, dstPool, dstImage, err := Image(string(destImageSpec))
	if err != nil {
		return err
	}

	dstIOCtx, err := conn.OpenIOContext(dstPool)
	if err != nil {
		return fmt.Errorf("failed to open destination pool (%s): %w", dstPool, err)
	}
	defer dstIOCtx.Destroy()
	dstIOCtx.SetNamespace(dstNamespace)

	imageOpts, err := rbdImageOptions(opts...)
	if err != nil {
		return fmt.Errorf("failed to build image options: %w", err)
	}
	defer imageOpts.Destroy()

	if err := cephrbd.MigrationPrepareImport(sourceSpec, dstIOCtx, dstImage, imageOpts); err != nil {
		return fmt.Errorf("failed to prepare import migration for image (%s): %w", destImageSpec, err)
	}
	return nil
}

// RbdMigrationExecute deep-copies source blocks to the migration target.
// Equivalent to: rbd migration execute <image-spec>
//
// When prog is non-nil, progress is reported by polling migration status while
// the blocking librbd call runs, matching native rbd CLI output.
func RbdMigrationExecute(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec, prog *Progress) error {
	if !imageSpec.Valid() {
		return fmt.Errorf("invalid image spec: %s", imageSpec)
	}

	namespaceName, poolName, imageName, err := Image(string(imageSpec))
	if err != nil {
		return err
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()
	ioctx.SetNamespace(namespaceName)

	if prog == nil {
		if err := cephrbd.MigrationExecute(ioctx, imageName); err != nil {
			return fmt.Errorf("failed to execute migration for image (%s): %w", imageSpec, err)
		}
		return nil
	}

	if err := migrationExecuteWithProgress(ctx, ioctx, imageName, imageSpec, prog); err != nil {
		return err
	}
	return nil
}

// RbdMigrationCommit commits a completed migration.
// Equivalent to: rbd migration commit <image-spec>
func RbdMigrationCommit(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec) error {
	return runMigrationImageOp(ctx, conn, imageSpec, "commit", cephrbd.MigrationCommit)
}

// RbdMigrationAbort aborts a prepared or executed migration.
// Equivalent to: rbd migration abort <image-spec>
func RbdMigrationAbort(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec) error {
	return runMigrationImageOp(ctx, conn, imageSpec, "abort", cephrbd.MigrationAbort)
}

// RbdMigrationStatus returns live-migration status for the target image.
// Equivalent to the migration section of: rbd status <image-spec>
func RbdMigrationStatus(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec) (*ImageStatusMigration, error) {
	if !imageSpec.Valid() {
		return nil, fmt.Errorf("invalid image spec: %s", imageSpec)
	}

	namespaceName, poolName, imageName, err := Image(string(imageSpec))
	if err != nil {
		return nil, err
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()
	ioctx.SetNamespace(namespaceName)

	image, err := cephrbd.OpenImageReadOnly(ioctx, imageName, cephrbd.NoSnapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	migration, err := readImageMigrationStatus(conn, ioctx, imageName, image)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration status for image (%s): %w", imageSpec, err)
	}
	return migration, nil
}

type migrationImageOp func(ioctx *cephrados.IOContext, imageName string) error

func runMigrationImageOp(
	ctx context.Context,
	conn *cephrados.Conn,
	imageSpec ImageSpec,
	op string,
	fn migrationImageOp,
) error {
	if !imageSpec.Valid() {
		return fmt.Errorf("invalid image spec: %s", imageSpec)
	}

	namespaceName, poolName, imageName, err := Image(string(imageSpec))
	if err != nil {
		return err
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()
	ioctx.SetNamespace(namespaceName)

	if err := fn(ioctx, imageName); err != nil {
		return fmt.Errorf("failed to %s migration for image (%s): %w", op, imageSpec, err)
	}
	return nil
}

func migrationProgressPercent(description string) (int, bool) {
	var pc int
	if _, err := fmt.Sscanf(strings.TrimSpace(description), "%d%% complete", &pc); err != nil {
		return 0, false
	}
	return pc, true
}

func migrationExecuteWithProgress(
	ctx context.Context,
	ioctx *cephrados.IOContext,
	imageName string,
	imageSpec ImageSpec,
	prog *Progress,
) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- cephrbd.MigrationExecute(ioctx, imageName)
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case err := <-errCh:
			if err != nil {
				prog.Fail()
				return fmt.Errorf("failed to execute migration for image (%s): %w", imageSpec, err)
			}
			prog.Finish()
			return nil
		case <-ctx.Done():
			prog.Fail()
			return ctx.Err()
		case <-ticker.C:
			raw, err := cephrbd.MigrationStatus(ioctx, imageName)
			if err != nil {
				continue
			}
			if pc, ok := migrationProgressPercent(raw.StateDescription); ok {
				prog.Update(pc)
			}
		}
	}
}
