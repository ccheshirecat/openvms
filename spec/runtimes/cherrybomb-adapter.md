# Cherrybomb Runtime Adapter for OVMS

## Overview

This document describes how to map OVMS manifest fields to Cherrybomb runtime start options. Cherrybomb is a hypothetical or specific VMM runtime; adapt the mapping to its CLI or API. Prefer implementing the OVMS Runtime Control API (OpenAPI at `spec/api/ovms-runtime.openapi.yaml`).

## Mapping Manifest to Cherrybomb Start

Use the manifest from examples/ubuntu-manifest.json as input. The adapter translates to Cherrybomb's expected flags (assuming CLI like hypr, but for Cherrybomb).

### Kernel and Initrd
- manifest.kernel.ref → Cherrybomb --kernel <path> (pull blob from OCI if needed).
- manifest.kernel.args → Cherrybomb --cmdline <args>.

### Disk Layers
- For each layer in manifest.diskLayers:
  - Pull blob to local path.
  - Apply in order: Cherrybomb --disk <path> (backing file for diffs).
  - Format: Map "qcow2" to Cherrybomb's format; fallback to raw.

### RAM Snapshot
- If manifest.ramSnapshot.ref present:
  - Pull blob, decompress (lz4).
  - Cherrybomb --ram-snapshot <path> (preload to tmpfs if hint=true).
  - mlock if mlock_required=true (use mlock syscall).

### Devices
- For each device in manifest.devices:
  - "ivshmem": Cherrybomb --ivshmem <name> --size <size> --addr <mmio_addr>.
  - Other types: Map to Cherrybomb device flags or error if unsupported.

### Runtime Hints
- manifest.runtimeHints.preferredRuntime: If "cherrybomb", proceed; else log warning.
- manifest.runtimeHints.coldStartTargetMs: Log for monitoring, not used in start.

### Example Start Command (CLI)

For the ubuntu-manifest.json:

1. Pull all refs using OCI client.
2. Decompress ram-snap if present.
3. Cherrybomb command:
   ```
   cherrybomb \
     --kernel /path/to/kernel \
     --initrd /path/to/initrd \
     --disk /path/to/base-layer.qcow2 \
     --disk /path/to/diff-1.qcow2 \
     --ram-snapshot /path/to/ram-snap \
     --ivshmem mmio0 --size 67108864 --addr 0x10000000 \
     --cmdline "console=ttyS0 root=/dev/vda1 rw"
   ```

## Implementation Notes

- Prefer HTTP client implementing OVMS Runtime Control API over UNIX socket.
- If using CLI, use Go exec.Command for Cherrybomb binary.
- Handle errors if format not supported.
- For integration with Huo: In VMM plugin, parse manifest, map to Cherrybomb, exec.
- Future: If Cherrybomb has API, use HTTP client instead of CLI.

## Limitations

- Assume Cherrybomb supports qcow2 diffs and IVSHMEM.
- No GPU or advanced devices in v0.1.
