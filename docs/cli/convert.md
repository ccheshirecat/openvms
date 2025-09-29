 # ovm convert (proposed)

 Non‑normative design for a neutral `ovm convert` subcommand that converts OCI/Docker images into bootable filesystem images for OVMS runtimes.

 Goals
 - Minimize friction moving from containers to microVM root filesystems.
 - Stay neutral: leverage standard tools (skopeo, umoci, mkfs.<fs>, mount/umount) when available.
 - Provide progress, error hints, and optional dual‑output (ext4 + squashfs).

 Sketch
 - Command: `ovm convert [flags] <oci-ref>`
 - Flags:
   - `--fs` ext4|xfs|btrfs (default: ext4)
   - `--size-buffer` MB (default: 50)
   - `--preallocate` (use fallocate)
   - `--dual-output` (also produce squashfs)
   - `--output` path (default: <name>.img)
   - `-v/--quiet/--no-color`

 Implementation Notes
 - Use `skopeo copy docker://ref oci:layout:latest` then `umoci unpack --image layout:latest`.
 - Compute size via `du -sk` of unpacked rootfs, add buffer, create sparse/preallocated image.
 - `mkfs.<fs>` with filesystem‑specific flags; mount via loop, copy files with progress; unmount and detach.
 - Optional: `mksquashfs` for dual output.
 - Requires root for mount operations; detect and hint when missing.

 Status
 - This document is a proposal; not normative. If implemented, it SHOULD live under the neutral OVMS CLI as an optional utility.
