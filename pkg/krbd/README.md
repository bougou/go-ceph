# krbd

The `krbd` subpackage maps and unmaps Ceph RBD images via the Linux kernel `rbd` driver (krbd), without requiring the external `rbd` CLI tool.

## Purpose

- Start RBD map/unmap requests programmatically.
- Use typed Go structs to describe mapping and unmapping requests.
- Keep map/unmap flows simple and consistent across environments.

## API overview

- `Image` describes an RBD target to map or unmap.
- `Options` carries optional krbd parameters such as credentials and mount behavior.
- `MapWriter()` and `UnmapWriter()` provide writers for map/unmap operations.
- `(*Image).Map(w)` sends a map request.
- `(*Image).Unmap(w)` sends an unmap request (supports force via `Options.Force`).
- `Devices()` and `Find()` help query currently mapped devices.

## Basic usage example

```go
w, err := krbd.MapWriter()
if err != nil {
    return err
}
defer w.Close()

img := krbd.Image{
    Monitors: []string{"10.0.0.1", "10.0.0.2"},
    Namespace: "",
    Pool:     "rbd",
    Image:    "demo-image",
    Options: &krbd.Options{
        Name:     "admin",
        Secret:   "<base64-key>",
        ReadOnly: true,
    },
}

if err := img.Map(w); err != nil {
    return err
}
```

For unmapping, use `UnmapWriter()` and call `Unmap()` with `Image{DevID: ...}`.
To force unmap, set `Options: &krbd.Options{Force: true}`.

## References

- This subpackage is inspired by and references the community project [`bensallen/rbd`](https://github.com/bensallen/rbd).
- It also follows the Linux kernel `rbd` `sysfs` ABI contract: [`Documentation/ABI/testing/sysfs-bus-rbd`](https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-bus-rbd).
- Option semantics are partially aligned with Ceph official docs for kernel rbd options: [`rbd(8) - kernel rbd (krbd) options`](https://docs.ceph.com/docs/master/man/8/rbd/#kernel-rbd-krbd-options).
