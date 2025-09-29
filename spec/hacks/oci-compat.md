# OCI Compatibility for OVMS

## Overview

OVMS is designed to be compatible with OCI registries and tools, allowing reuse of existing infrastructure for distribution. This document outlines how OVMS artifacts map to OCI concepts and any hacks/workarounds for compatibility. Aligns with OCI 1.1 fields (`artifactType`, `subject`) and Referrers API.

## Mapping to OCI

- **Manifest**: OVMS manifest.json is the primary artifact. Media type: `application/vnd.ovms.manifest.v1+json`. For OCI 1.1, set the manifest's `artifactType` to this value. Related artifacts (e.g., signatures, snapshots, SBOMs) use `subject` to reference the base manifest and are discoverable via the Referrers API.
- **Blobs**: Kernel, initrd, disk layers, RAM snapshots are OCI blobs.
  - Kernel: Media type `application/vnd.ovms.kernel.v1`.
  - Initrd: `application/vnd.ovms.initrd.v1`.
  - Disk layer: `application/vnd.oci.image.layer.v1.tar` with annotation `ovms.format=qcow2|raw`.
  - RAM snapshot: `application/vnd.ovms.ramsnap.v1+lz4`.
- **Registry Push/Pull**: Use oras or skopeo with custom media types. Example:
  ```
  oras push myregistry.ovms/ubuntu:24.04 /path/to/ovm --media-type application/vnd.ovms.manifest.v1+json --artifact-type application/vnd.ovms.manifest.v1+json
  oras pull myregistry.ovms/ubuntu:24.04
  ```
- **Local Storage**: Use OCI image layout for local cache (index.json, blobs/sha256/, oci-layout).

## Hacks and Workarounds

- **Annotation for Format**: Since OCI layers are tar, add annotation `ovms.format=qcow2` to disk layer descriptors. Runtimes parse annotations to determine format.
- **RAM Snapshot as Config**: Treat RAM snapshot as a special "config" blob if no config.json exists; use media type to identify.
- **Multi-Arch Support**: Use OCI multi-arch manifests for different platforms (x86_64, arm64). Runtime selects based on manifest.platform.arch.
- **Provenance**: Use Cosign for signing OVMS manifests. Signatures are attached as artifacts that set `subject` to the OVMS manifest and are discoverable via Referrers.
- **Fallback to OCI Containers**: If OVMS manifest has "containerMode": true, treat disk layer as OCI container image and boot VM with container runtime (e.g., runc inside QEMU).
- **Digest Verification**: Always verify blob digests on pull to ensure integrity.
- **Tool Compatibility**: 
  - Skopeo: Supports custom media types for push/pull.
  - Docker: Use `docker load` for local, but for OVMS, use oras for custom types.
  - Workaround for Docker: Export as OCI tar, then use oras to push with media types.

## Implementation Notes (generic)

- Use ORAS-compatible clients to pull manifest and blobs.
- Parse manifest.json to get refs and annotations.
- Fetch blobs via OCI APIs; verify digests and signatures when present.
- Apply disk layer chains via qcow2 backing files or equivalent.
- For RAM snapshots, decompress and load per runtime capabilities.

This ensures OVMS works with existing OCI tools while extending for VM artifacts.
