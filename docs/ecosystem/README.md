# OVMS Ecosystem (Non‑Normative)

This page lists community tools that can help users adopt OVMS. These references are non‑normative and optional; the specification remains tooling‑agnostic.

Categories:

- Filesystem image conversion (OCI → bootable FS images)
  - fsify — Convert Docker/OCI images into bootable filesystem images (ext4/xfs/btrfs; optional squashfs). Open source and independent of any orchestrator.
    • Repository: https://github.com/ccheshirecat/fsify
    • Typical usage: `sudo fsify nginx:latest` → `nginx-latest.img`

Notes:
- Listing here does not imply endorsement. Implementations can provide similar functionality directly (e.g., an `ovm convert` subcommand) or via other tools.
- Tools SHOULD document required privileges and dependencies (mkfs, mount, skopeo, umoci, etc.).
