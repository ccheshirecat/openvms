# Firecracker Runtime Adapter for OVMS

## Overview
Maps OVMS manifest fields to Firecracker VMM. Prefer implementing the OVMS Runtime Control API (see `spec/api/ovms-runtime.openapi.yaml`) via a local shim that translates API calls into Firecracker's API/CLI.

## Mapping
- Kernel: manifest.kernel.ref → kernel_image_path; args → boot_args
- Initrd: manifest.initrd.ref → initrd_path
- Disks: manifest.diskLayers → merge/apply into a single block device (qcow2 backing chain) → block_devices[].
- RAM Snapshot: manifest.ramSnapshot → Firecracker snapshot/restore (if available) or shim for fast-boot preload.
- Devices: ivshmem not natively supported; use vsock or MMDS alternatives where applicable.

## Start Flow (Shim)
1. Pull blobs via ORAS.
2. Build rootfs qcow2 chain; expose as block device.
3. Launch Firecracker with kernel/initrd/boot_args and block device.
4. Expose control socket for status/logs.

## Example
```bash
firecracker \
  --kernel /path/kernel \
  --initrd /path/initrd \
  --boot-args "console=ttyS0 root=/dev/vda1 rw" \
  --rootfs /path/rootfs.qcow2 \
  --tap tap0
```

## Notes
- For snapshot fast-boot, prefer Firecracker native snapshot/restore if available.
- Align published snapshot artifact with OVMS RAM snapshot guidance.