package rbd

import (
	"context"
	"fmt"
	"strconv"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

// ImageStatus is the result of "rbd status".
type ImageStatus struct {
	Watchers  []cephrbd.ImageWatcher `json:"watchers"`
	Migration *ImageStatusMigration  `json:"migration,omitempty"`
}

// ImageStatusMigration contains migration details when the image has
// RBD_FEATURE_MIGRATING set.
type ImageStatusMigration struct {
	// SourceSpec is the JSON source-spec for import-only migrations.
	SourceSpec string `json:"source_spec,omitempty"`
	// Source is the formatted source (native image ref or import source-spec JSON).
	Source string `json:"source"`
	// Destination is the formatted destination image ref.
	Destination string `json:"destination"`
	// State is the librbd migration state.
	State cephrbd.MigrationImageState `json:"state"`
	// StateDescription provides extra detail (e.g. execute progress).
	StateDescription string `json:"state_description,omitempty"`
}

func migrationStateName(state cephrbd.MigrationImageState) string {
	switch state {
	case cephrbd.MigrationImageError:
		return "error"
	case cephrbd.MigrationImagePreparing:
		return "preparing"
	case cephrbd.MigrationImagePrepared:
		return "prepared"
	case cephrbd.MigrationImageExecuting:
		return "executing"
	case cephrbd.MigrationImageExecuted:
		return "executed"
	case cephrbd.MigrationImageAborting:
		return "aborting"
	default:
		return "unknown"
	}
}

// MigrationStateLine returns the migration state string as shown by "rbd status".
func MigrationStateLine(m *ImageStatusMigration) string {
	if m == nil {
		return ""
	}
	line := migrationStateName(m.State)
	if m.StateDescription != "" {
		line += " (" + m.StateDescription + ")"
	}
	return line
}

func poolNameByID(conn *cephrados.Conn, poolID int) string {
	if poolID < 0 {
		return ""
	}
	name, err := conn.GetPoolByID(int64(poolID))
	if err != nil {
		return strconv.Itoa(poolID)
	}
	return name
}

func formatPoolImageRef(poolName, namespace, imageName, imageID string) string {
	ref := poolName
	if namespace != "" {
		ref += "/" + namespace
	}
	ref += "/" + imageName
	if imageID != "" {
		ref += " (" + imageID + ")"
	}
	return ref
}

func buildImageStatusMigration(
	conn *cephrados.Conn,
	raw *cephrbd.MigrationImageStatus,
	sourceSpec string,
) *ImageStatusMigration {
	mig := &ImageStatusMigration{
		SourceSpec:       sourceSpec,
		State:            raw.State,
		StateDescription: raw.StateDescription,
	}

	if sourceSpec != "" {
		mig.Source = sourceSpec
	} else if raw.SourcePoolID >= 0 {
		srcPool := poolNameByID(conn, raw.SourcePoolID)
		mig.Source = formatPoolImageRef(
			srcPool,
			raw.SourcePoolNamespace,
			raw.SourceImageName,
			raw.SourceImageID,
		)
	}

	dstPool := poolNameByID(conn, raw.DestPoolID)
	mig.Destination = formatPoolImageRef(
		dstPool,
		raw.DestPoolNamespace,
		raw.DestImageName,
		raw.DestImageID,
	)
	return mig
}

func readImageMigrationStatus(
	conn *cephrados.Conn,
	ioctx *cephrados.IOContext,
	imageName string,
	image *cephrbd.Image,
) (*ImageStatusMigration, error) {
	features, err := image.GetFeatures()
	if err != nil {
		return nil, fmt.Errorf("failed to get image features: %w", err)
	}
	if features&cephrbd.FeatureMigrating == 0 {
		return nil, nil
	}

	raw, err := cephrbd.MigrationStatus(ioctx, imageName)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration status: %w", err)
	}

	var sourceSpec string
	if raw.SourcePoolID < 0 {
		// Import-only migrations require rbd_get_migration_source_spec, which is
		// not yet exposed by github.com/ceph/go-ceph.
		sourceSpec = ""
	}

	return buildImageStatusMigration(conn, raw, sourceSpec), nil
}

// RbdStatus returns watchers and migration status for an image or snapshot.
// Equivalent to: rbd status <image-or-snap-spec>
func RbdStatus(ctx context.Context, conn *cephrados.Conn, imageOrSnapSpec string) (*ImageStatus, error) {
	namespaceName, poolName, imageName, snapshotName, err := ImageOrSnap(imageOrSnapSpec)
	if err != nil {
		return nil, err
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to open pool (%s): %w", poolName, err)
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	snapName := snapshotName
	if snapName == "" {
		snapName = cephrbd.NoSnapshot
	}

	// OpenImageReadOnly avoids registering this client as a watcher.
	image, err := cephrbd.OpenImageReadOnly(ioctx, imageName, snapName)
	if err != nil {
		return nil, fmt.Errorf("failed to open image (%s): %w", imageName, err)
	}
	defer image.Close()

	watchers, err := image.ListWatchers()
	if err != nil {
		return nil, fmt.Errorf("failed to get watchers: %w", err)
	}

	status := &ImageStatus{Watchers: watchers}

	migration, err := readImageMigrationStatus(conn, ioctx, imageName, image)
	if err != nil {
		return nil, err
	}
	status.Migration = migration

	return status, nil
}
