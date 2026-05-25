# go-ceph

Opinionated extensions around the official `github.com/ceph/go-ceph`, currently focused on practical RBD workflows.

## What this repository is for

This project builds on top of the official Go Ceph bindings and adds:

- Higher-level RBD helper APIs for common operations.
- A workflow-oriented layer for scripts and automation.

## How it differs from the official `go-ceph`

- **Official `go-ceph`**: low-level, generic Go bindings for Ceph APIs.
- **This repository**: adds convenience wrappers for common RBD operations, so app code needs less boilerplate.

## Scope and roadmap

- **Current scope**: mostly RBD-related helpers.
- **Planned expansion**: CephFS, RGW, and more Ceph capabilities over time.

## Stability

- APIs are evolving and may change as coverage expands.
- Pin a module version in production and review changelogs before upgrades.

## Quick usage (library)

```go
package main

import (
	"context"
	"log"

	"github.com/bougou/go-ceph/pkg/rados"
	"github.com/bougou/go-ceph/pkg/rbd"
)

func main() {
	ctx := context.Background()

	conn, err := rados.NewRadosConn("", false)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := conn.RbdCreate(ctx, rbd.ImageSpec("pool/image"), 1<<30); err != nil {
		log.Fatal(err)
	}

	info, err := conn.RbdInfo(ctx, rbd.ImageSpec("pool/image"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println(info.String())
}
```
