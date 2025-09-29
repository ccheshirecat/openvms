# Open Virtual Machine Spec (OVMS) v0.1

## Mission
OVMS defines a portable specification for packaging, distributing, and running full-virtual-machine artifacts (disk layers, kernel/initrd, and optional RAM snapshots) with a small runtime contract. This enables any runtime to be a drop-in for orchestrators like Huo, supporting fast boot, layering, and OCI compatibility for distribution.

## Quickstart
1. **Create Manifest**: Use the JSON schema in schemas/manifest.v1.json to define your VM artifact.
2. **Build Artifact**: Package kernel, initrd, disk layers, and RAM snapshot as OCI blobs (use tools like skopeo for OCI push).
3. **Push to Registry**: Push to OCI registry (e.g., Docker Hub, private).
4. **Pull and Run**: Use CLI `ovm pull <ref>` to fetch, then `ovm run <ref>` to start.
5. **Snapshot**: `ovm snapshot <instance> -t <tag>` to create RAM snapshot.
6. **Integrate with Huo**: Use RuntimePlugin in Huo to dispatch OVMS workloads to VMM plugins (e.g., Hypr adapter).

## Overview
OVMS reuses OCI for distribution (artifacts use OCI image manifests with custom media types and OCI 1.1 fields such as `artifactType` and `subject`). Artifacts are layered for efficiency: base disk + diffs + optional RAM snapshot for <10ms boots. Runtimes expose a minimal HTTP API for control. Huo integrates as the reference orchestrator via RuntimePlugin interface.

See rfc/ovms-v0.1.md for full spec and media types (e.g., `application/vnd.ovms.manifest.v1+json`).

## Components
- **Manifest**: JSON describing artifact (see examples/ubuntu-manifest.json).
- **Runtime Contract**: HTTP endpoints for start/stop/snapshot/status (see rfc/ovms-v0.1.md).
- **CLI**: `ovm` tool for pull/run/snapshot (see cli/ovm).
- **Adapters**: Mappings for runtimes like Cherrybomb (see runtimes/cherrybomb-adapter.md).
- **Hacks**: OCI compatibility notes (see hacks/oci-compat.md).

## Security
- Sign manifests with Cosign (see docs/signing/README.md). Use OCI Referrers and `subject` to relate signatures to manifests.
- Runtimes use namespaces, seccomp for isolation.
- RAM snapshots opt-in with mlock for latency guarantees.

## Limitations v0.1
- Linux x86_64 only.
- BIOS boot (UEFI v0.2).
- No GPU; basic devices (IVSHMEM).

For Huo integration, see Huo's RuntimePlugin interface in pkg/plugin.