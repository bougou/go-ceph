# krbd

The `krbd` subpackage maps and unmaps Ceph RBD images through the Linux kernel `rbd` driver (kernel rbd).

It is designed for cases where Go code should talk directly to `sysfs` and reuse the kernel `krbd` interface, instead of depending on the external `rbd` CLI tool.

## Purpose

- Start RBD map/unmap requests programmatically.
- Assemble Ceph monitor, pool, image, snapshot, and krbd options into the format expected by the kernel.
- Automatically pick a valid `sysfs` write endpoint across kernel/distribution path differences.

## Implementation

The current `krbd` implementation is built around `sysfs`, with four main parts:

1. **Request modeling**
   - `Image` represents the core fields for a map/unmap operation (monitors, pool, image, snapshot, dev id, etc.).
   - `Options` represents optional krbd parameters (such as `name`, `secret`, `read-only`, `queue_depth`, etc.).

2. **Parameter encoding**
   - `Image.MarshalText()` encodes image fields into the string format expected by the kernel `add` interface.
   - `Options.MarshalText()` converts struct fields (based on `krbd` tags) into comma-separated options; boolean options are emitted by key presence.
   - `Options.UnmarshalText()` parses a comma-separated options string back into `Options`.

3. **Map/Unmap execution**
   - `Image.Map(w io.Writer)` writes map requests to `add`-style interfaces.
   - `Image.Unmap(w io.Writer)` writes unmap requests to `remove`-style interfaces (supports `force`).
   - `MapWriter()` / `UnmapWriter()` open the appropriate `sysfs` writer.

4. **Kernel parameter compatibility**
   - Reads `/sys/module/rbd/parameters/single_major` to detect whether `single_major` is enabled.
   - Based on that parameter, it tries in this order:
     - `add_single_major` / `remove_single_major`
     - fallback to `add` / `remove`
   - Supports both `/sys/bus/rbd` and `/sys/bus/rbd/devices` path layouts.

## Basic usage example

```go
w, err := krbd.MapWriter()
if err != nil {
    return err
}
defer w.Close()

img := krbd.Image{
    Monitors: []string{"10.0.0.1:6789", "10.0.0.2:6789"},
    Pool:     "rbd",
    Image:    "demo-image",
    Snapshot: "-",
    Options: &krbd.Options{
        Name:   "client.admin",
        Secret: "<base64-key>",
        ReadOnly: true,
    },
}

if err := img.Map(w); err != nil {
    return err
}
```

For unmapping, use `UnmapWriter()` and call `Unmap()` with `Image{DevID: ...}`.

## References and source notes

- This subpackage is inspired by and references the community project [`bensallen/rbd`](https://github.com/bensallen/rbd).
- It also follows the Linux kernel `rbd` `sysfs` ABI contract: [`Documentation/ABI/testing/sysfs-bus-rbd`](https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-bus-rbd).
- Option semantics are partially aligned with Ceph official docs for kernel rbd options: [`rbd(8) - kernel rbd (krbd) options`](https://docs.ceph.com/docs/master/man/8/rbd/#kernel-rbd-krbd-options).