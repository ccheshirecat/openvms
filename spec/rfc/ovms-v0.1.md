# OVMS v0.1 — RFC Skeleton (fast-track)

## Purpose

Define a portable Open Virtual Machine Spec (OVMS) for packaging, distributing, and running full-virtual-machine artifacts (disk layers, kernel/initrd, and optional RAM snapshots) with a small runtime contract so any runtime (VMM implementation) can be a drop-in for orchestrators. This spec enables VMM-agnostic orchestration by allowing workloads to specify OVMS manifests in metadata, with adapters dispatching to the appropriate runtime implementation.

## Design goals
- OCI-compatible distribution (reuse OCI distribution API and registries for push/pull).
- Layered images: base disk + diffs + optional RAM snapshot layer for efficient storage and fast boots.
- Manifest-first: single manifest.json describes runtime needs (kernel, disks, RAM snapshot, devices).
- Runtime contract: minimal HTTP/gRPC API for start/stop/snapshot/inject/status, implemented by runtimes.
- Pluggable acceleration: IVSHMEM, snapshot preloading, tmpfs hints for low-latency.
- Runtime-agnostic: Runtimes implement the contract; orchestrators dispatch via adapters.

## 1 — Manifest format (manifest.json) — canonical example

{
  "schemaVersion": 1,
  "mediaType": "application/vnd.ovms.manifest.v1+json",
  "name": "example/ubuntu-base",
  "version": "24.04",
  "kernel": {
    "ref": "sha256:kernel-digest (OCI blob ref or external URL)",
    "args": "console=ttyS0 root=/dev/vda1 rw"
  },
  "initrd": { "ref": "sha256:initrd-digest" },
  "diskLayers": [
    { "ref": "sha256:base-layer", "format": "qcow2", "size": 2147483648 },
    { "ref": "sha256:diff-1", "format": "qcow2", "size": 33554432 }
  ],
  "ramSnapshot": {
    "ref": "sha256:ram-snap",
    "compression": "lz4",
    "preload_hint": true,
    "mlock_required": true
  },
  "devices": [
    { "type": "ivshmem", "name": "mmio0", "size": 67108864, "mmio_addr": "0x10000000" }
  ],
  "metadata": {
    "author": "ovms",
    "created": "2025-09-09T00:00:00Z",
    "platform": { "arch": "x86_64", "uefi": false }
  },
  "runtimeHints": {
    "preferredRuntime": ["generic"],
    "coldStartTargetMs": 10
  }
}

- Notes
- ref = OCI blob digest or registry reference (e.g., myregistry.com/ovms/ubuntu:24.04/kernel). Clients may pull using ORAS.
- diskLayers order: base first, diffs applied top-down (copy-on-write for efficiency).
- ramSnapshot is optional; if present, runtime may preload to tmpfs for sub-ms boot if hint=true.
- devices encodes special devices (IVSHMEM for shared memory, etc.) in declarative way; runtime maps to flags.
  

## 2 — Media Types & OCI mapping
- Manifest media type: application/vnd.ovms.manifest.v1+json (parsed via JSON schema). For OCI v1.1, set artifactType to this value and use subject for relationships.
- RAM snapshot blob media type: application/vnd.ovms.ramsnap.v1+lz4 (decompress with lz4).
- Disk layer media type: application/vnd.oci.image.layer.v1.tar with annotation ovms.format=qcow2/raw/ovfms.
- Kernel blob type: application/vnd.ovms.kernel.v1.
- Initrd blob type: application/vnd.ovms.initrd.v1.
- Use OCI distribution API (v2) for push/pull with ORAS. Prefer OCI 1.1 fields (artifactType, subject/referrers) for relationships.

## 3 — Layering model / snapshot model
- Disk layers: Block-level diffs (qcow2 backing-file style) for dedupe and lazy fetch. Adapters apply layers before start.
- RAM snapshot:
  - Blob of compressed memory pages + metadata (CPU registers, device state).
  - Compression: lz4 default; future zstd.
  - Runtime preloads to tmpfs if preload_hint=true, mlock if mlock_required=true for latency guarantees.
  - Applying layers: Runtimes apply disk layers to ephemeral/persistent store; if ramSnapshot present and supported, bypass disk boot for <10ms cold starts.
  

## 4 — Runtime contract (HTTP/gRPC surface)
Runtimes expose a control API (local UNIX socket). Minimal REST endpoints (gRPC optional for future).

HTTP endpoints (suggested minimal surface):
- POST /start — body: {"manifest": "<path|OCI-ref|tar>", "overrides": {...}} → returns {"instance_id":"uuid", "status":"starting"}.
- POST /stop — body: {"instance_id":"uuid"} → {"status":"stopping"}.
- POST /snapshot — body: {"instance_id":"uuid", "tag":"fastboot"} → returns {"snapshot_ref":"sha256:..."}.
- POST /inject — body: {"instance_id":"uuid", "content": {"mmio": {...}}} → {"status":"ok"}.
- GET /status/{instance_id} → returns JSON InstanceInfo with metrics.
- GET /logs/{instance_id} → streaming logs (or WS).

gRPC alternative: Define protobuf with RPCs: Start(ManifestRef) returns Instance, Stop(InstanceID), Snapshot(InstanceID) returns SnapshotRef, Status(InstanceID) returns Status.

Rationale: Small surface for quick runtime implementation.

OpenAPI: A normative OpenAPI document is provided at `spec/api/ovms-runtime.openapi.yaml` describing request/response schemas for `/start`, `/stop`, `/snapshot`, `/status/{instance_id}`, and `/logs/{instance_id}`.

## 5 — CLI UX (ovm)
Basic commands for standalone use (skeleton in cli/ovm/main.go):
- ovm pull example/ubuntu-base:24.04 — pulls manifest + blobs to local cache.
- ovm run example/llm:fastboot --runtime=<impl> --memory=16G — starts instance, prints ID.
- ovm snapshot <instance> -t example/llm:fastboot — creates RAM snapshot, pushes to registry.
- ovm push example/llm:fastboot — pushes manifest + blobs.
- ovm inspect example/... — shows manifest.
- ovm ls — lists local artifacts.

CLI uses OCI distribution for push/pull; local cache in OCI layout extended with ovms.blobs/ for RAM snapshots.

## 6 — Security & policy
- Runtimes honor Linux security: namespaces, seccomp, user namespaces for management processes.
- Sign manifests with Cosign; implementations SHOULD verify on pull. OVMS artifacts SHOULD use OCI 1.1 relationships:
  - `artifactType` = `application/vnd.ovms.manifest.v1+json` for OVMS manifests.
  - Use `subject` and the Referrers API to relate signatures, SBOMs, and snapshots to the base manifest.
  - Verification policy MAY require trusted signatures before runtime admission.
- Access control: Registry ACLs + runtime-local policies (e.g., disallow unsigned ramSnapshot).
  

## 7 — Backwards compatibility / migration
- OVMS artifacts wrap OCI container images (disk layer runs container runtime in VM) for compatibility.
- Keep OCI registry semantics for adoption.
- Migration from legacy runtime arguments: Convert to manifest, push as OVMS.

## 8 — Minimal file layout (OVMS artifact when downloaded)
/ovm/
  manifest.json
  blobs/
    sha256-<kernel>
    sha256-<initrd>
    sha256-<disk-layer-1>
    sha256-<ram-snap>
  metadata/
    provenance.json (Cosign signature)

## 9 — v0.1 Limitations (explicit)
- Linux x86_64 only.
- BIOS boot (UEFI v0.2).
- No GPU passthrough (v0.2).
- Complex PCI state save/restore deferred.

 
