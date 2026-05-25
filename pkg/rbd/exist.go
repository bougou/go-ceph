package rbd

import (
	"context"
	"fmt"

	cephrados "github.com/ceph/go-ceph/rados"
	cephrbd "github.com/ceph/go-ceph/rbd"
)

func RbdExist(ctx context.Context, conn *cephrados.Conn, imageSpec ImageSpec) (exist bool, err error) {
	namespaceName, poolName, imageName, err := Image(string(imageSpec))
	if err != nil {
		return
	}

	ioctx, err := conn.OpenIOContext(poolName)
	if err != nil {
		err = fmt.Errorf("failed to open pool (%s): %w", poolName, err)
		return
	}
	defer ioctx.Destroy()

	ioctx.SetNamespace(namespaceName)

	image, err := cephrbd.OpenImageReadOnly(ioctx, imageName, cephrbd.NoSnapshot)
	if err != nil {
		if isErrNotFound(err) {
			err = nil
			return
		}
		err = fmt.Errorf("failed to open image (%s): %w", imageName, err)
		return
	}
	defer image.Close()

	exist = true
	return
}
