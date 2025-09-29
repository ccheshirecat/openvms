# Runtime Adapter Guidance for OVMS

This document provides generic guidance for mapping OVMS manifests to a virtual machine monitor (VMM) runtime. It avoids naming specific tools to keep the specification impartial.

## Kernel and Initrd
- manifest.kernel.ref → runtime kernel path (pull blob from OCI if needed).
- manifest.kernel.args → runtime boot/append arguments.

## Disk Layers
- For each layer in manifest.diskLayers:
  - Pull blob to local path.
  - Apply in order (base first, diffs on top) using backing files or equivalent.
  - Map format (e.g., qcow2/raw) to the runtime’s supported disk format.

## RAM Snapshot
- If manifest.ramSnapshot.ref present:
  - Pull blob, decompress (e.g., lz4).
  - Load/restore according to runtime capability; optionally preload into tmpfs if `preload_hint=true`.
  - mlock if `mlock_required=true` and supported by the runtime.

## Devices
- For each device in manifest.devices:
  - Example: "ivshmem" → map to shared memory device per runtime syntax (name/size/mmio address).
  - Return clear errors for unsupported device types.

## Runtime Hints
- `runtimeHints.preferredRuntime` is advisory; adapters may ignore if not applicable.
- `coldStartTargetMs` is informational for monitoring and tuning.

## Example Start Flow (Generic)
1. Pull all refs using an OCI client.
2. Assemble disk chain and optional RAM snapshot.
3. Launch the runtime with kernel/initrd/boot args and block devices.
4. Expose a local control API implementing `spec/api/ovms-runtime.openapi.yaml` for status/logs/snapshot/stop.

## Notes
- Prefer an HTTP client over UNIX socket implementing the OVMS Runtime Control API.
- If using a CLI-based runtime, a thin shim can translate API calls to process invocations.
- Handle unsupported features gracefully and document limitations.
