# QEMU Runtime Adapter for OVMS

## Overview
Maps OVMS manifest to QEMU CLI. Prefer using an OVMS runtime shim implementing `spec/api/ovms-runtime.openapi.yaml` that constructs the QEMU command line and manages lifecycle.

## Mapping
- Kernel: -kernel <kernel> -append "<args>"
- Initrd: -initrd <initrd>
- Disks: -drive file=<qcow2>,if=virtio,format=qcow2 (apply backing chain)
- RAM Snapshot: loadvm or migration-based restore if applicable; otherwise preload into tmpfs and boot.
- Devices: ivshmem → -device ivshmem-plain,size=<size>,addr=<mmio>

## Start Flow
1. Pull blobs via ORAS; assemble qcow2 chain.
2. Compose QEMU args from manifest; start process and capture pid/logs.
3. Expose control API for status/stop/snapshot.

## Example
```bash
qemu-system-x86_64 \
  -enable-kvm -m 4096 -smp 2 \
  -kernel /path/kernel -initrd /path/initrd \
  -append "console=ttyS0 root=/dev/vda1 rw" \
  -drive file=/path/rootfs.qcow2,if=virtio,format=qcow2 \
  -nographic
```

## Notes
- Prefer virtio devices; align with OVMS device hints.
- Snapshot semantics vary; treat OVMS RAM snapshot blob as implementation-defined for v0.1.