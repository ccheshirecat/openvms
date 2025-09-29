# Open Virtual Machine Spec (OVMS): Current State, Prior Art, and Proposed Plan

## Executive Summary

OVMS aims to define a portable, OCI‑compatible way to package, distribute, and run full VM artifacts (kernel/initrd, layered disks, optional RAM snapshots) with a minimal runtime contract. The current repository already contains a v0.1 RFC skeleton, a JSON Schema for manifests, examples, OCI compatibility notes, a CLI scaffold, and a runtime adapter mapping. Building on industry standards (OCI Image/Distribution, ORAS, Sigstore/Cosign) and lessons from established VM and bytecode specs (JVM, WebAssembly, ECMA‑335/CLI, EVM), this plan proposes a staged path to a ratified OVMS v0.1 specification with a conformance suite and reference implementations, followed by v0.2 enhancements for UEFI, multi‑arch, and richer snapshot semantics.

Top priorities:
- Finalize v0.1 normative spec text, media types, and JSON Schema aligned with OCI Image/Distribution and Referrers/Artifacts v1.1 [1][2][3][4][18][19].
- Define minimal runtime control API and map to at least two VMMs; provide a working adapter per runtime.
- Ship a conformance test suite (schema, media types, registry behavior, runtime contract) and a signed artifact story using Cosign [9].
- Publish reference tooling (ovm CLI via ORAS) and examples.

## What We Have Today (Repo Review)

Artifacts discovered in this repository:
- Spec draft: `spec/rfc/ovms-v0.1.md` — structure, goals, manifest example, runtime contract, media types, security, limitations.
- JSON Schema: `spec/schemas/manifest.v1.json` — validates fields such as kernel/initrd refs, disk layers, optional RAM snapshot, devices, metadata, and runtimeHints.
- Examples: `spec/examples/ubuntu-manifest.json`.
- OCI compatibility: `spec/hacks/oci-compat.md` — mapping to OCI manifests/blobs and notes about annotations and multi‑arch.
- Runtime adapter notes: `spec/runtimes/cherrybomb-adapter.md`.
- CLI skeleton: `spec/cli/ovm/main.go` (uses ORAS; scaffolding for pull/run/etc.).

Gaps to address:
- Normative media types and alignment with OCI v1.1 fields (`artifactType`, `subject`, Referrers API) [3][18][19].
- Precise, testable runtime API definitions (schemas/IDL and conformance checks).
- Formal conformance across distribution (OCI registry behavior), artifact format, and runtime control.
- Security requirements (signing/verification) and supply chain alignment (Cosign/Referrers) [9][18][19].

## Prior Art and Lessons

Distribution and artifact model
- OCI Distribution Spec and API v2 [1][2].
- OCI Image Spec: media types and image layout [3][4][5].
- ORAS as generic client for non‑container artifacts (custom media types, config, annotations) [6][7][8].
- OCI v1.1 additions: `artifactType`, `subject`, and Referrers API enable first‑class non‑image artifacts and relationships (signatures, SBOMs, snapshots) [18][19].
- Signing: Sigstore/Cosign supports signing arbitrary OCI artifacts and indices [9].

VM formats and performance constructs
- QCOW2 format (backing files, copy‑on‑write layering) [10].
- IVSHMEM device spec for shared memory/low‑latency communication [11].
- Firecracker/microVM snapshotting and fast restore (cold‑start reductions to sub‑10ms) [12][13].

Specification organization patterns
- WebAssembly Core Specification: clear layering of ISA, binary/text formats, validation, and conformance mindset [14].
- JVM Specification: structured sections for class file format, execution, exceptions, semantics; enduring reference across versions [15].
- ECMA‑335/CLI: partitioned spec (architecture, metadata, IL, profiles, binary formats) useful for modular spec drafting [16].
- Ethereum Yellow Paper: formal, math‑heavy approach; demonstrates value of precise semantics and test vectors for interoperability [17].

Key takeaways for OVMS
- Reuse OCI registry and artifact infrastructure; register media types; leverage `artifactType` and `subject` for relationships (e.g., snapshots, attestations) [3][18][19].
- Provide precise, testable definitions (schemas, protocol) and a conformance suite.
- Design for multiple runtimes via a minimal contract; publish adapter mappings.
- Prioritize fast‑boot via layered disks and optional RAM snapshots with security and uniqueness considerations [10][12].

## Scope and Goals for OVMS v0.1

In scope (v0.1)
- Distribution as OCI artifacts (manifest + blobs), with normative media types and layout [1][3][4].
- Manifest JSON schema v1 and canonical example (kernel/initrd refs, disk layers, devices, metadata, optional RAM snapshot).
- Runtime control API (local HTTP over Unix socket) for start/stop/snapshot/status/logs.
- Security baseline: Cosign‑signed artifacts; digest verification on pull; optional policy for snapshot acceptance [9].
- Platform baseline: Linux x86_64, BIOS boot; QCOW2 and RAW disk support; IVSHMEM optional [10][11].

Out of scope (defer to v0.2+)
- UEFI, multi‑arch matrices, GPU and advanced devices, rich migration semantics.
- Formal RAM snapshot format standardization across VMMs; start with opaque blob per runtime, define metadata contract first [12].

## Proposed Normative Elements (v0.1)

1) Media types and OCI alignment
- Manifest: `application/vnd.ovms.manifest.v1+json` (consider alias under org vendor namespace initially; migrate to vnd.ovms upon registration) [3].
- Kernel blob: `application/vnd.ovms.kernel.v1`.
- Initrd blob: `application/vnd.ovms.initrd.v1`.
- Disk layer blob: `application/vnd.oci.image.layer.v1.tar` with annotation `ovms.format=qcow2|raw` [3].
- RAM snapshot blob: `application/vnd.ovms.ramsnap.v1+lz4` (or `+zstd` when supported).
- OVMS manifest SHOULD set `artifactType: application/vnd.ovms.manifest.v1+json`; snapshot or attestation objects SHOULD set `subject` to the referenced manifest/instance [18][19].

2) Manifest schema and example (excerpt)
```json
{
  "schemaVersion": 1,
  "mediaType": "application/vnd.ovms.manifest.v1+json",
  "name": "example/ubuntu-base",
  "version": "24.04",
  "kernel": { "ref": "sha256:<digest>", "args": "console=ttyS0 root=/dev/vda1 rw" },
  "initrd": { "ref": "sha256:<digest>" },
  "diskLayers": [
    { "ref": "sha256:<base>", "format": "qcow2", "size": 2147483648 },
    { "ref": "sha256:<diff1>", "format": "qcow2", "size": 33554432 }
  ],
  "ramSnapshot": { "ref": "sha256:<snap>", "compression": "lz4", "preload_hint": true, "mlock_required": true },
  "devices": [ { "type": "ivshmem", "name": "mmio0", "size": 67108864, "mmio_addr": "0x10000000" } ],
  "metadata": { "author": "ovms", "created": "2025-09-29T00:00:00Z", "platform": { "arch": "x86_64", "uefi": false } },
  "runtimeHints": { "preferredRuntime": ["firecracker", "qemu"], "coldStartTargetMs": 10 }
}
```

3) Runtime control API (local HTTP over Unix socket)
- POST `/start` → `{ "manifest": "<oci-ref|path>", "overrides": {...} }` → `{ "instance_id": "uuid", "status": "starting" }`
- POST `/stop` → `{ "instance_id": "uuid" }` → `{ "status": "stopping" }`
- POST `/snapshot` → `{ "instance_id": "uuid", "tag": "fastboot" }` → `{ "snapshot_ref": "sha256:..." }`
- GET `/status/{instance_id}` → `InstanceInfo` (state, timings, resources)
- GET `/logs/{instance_id}` → stream

Notes:
- Socket permissioning and auth are implementation‑specific; non‑goals for v0.1 but MUST be isolated per‑host policy.
- gRPC schema may be provided as a non‑normative alternative (future).

4) Security and supply chain
- MUST verify blob digests; SHOULD verify Cosign signatures when present; MAY require policy to run only signed artifacts [9].
- SHOULD support OCI Referrers graph for signatures/SBOMs/snapshots via `subject` discovery [18][19].

5) Distribution guidance
- Use ORAS or compatible clients to push/pull artifacts with custom media types [6][7][8].
- Prefer `artifactType` over legacy `config.mediaType` typing; when absent, fall back per OCI guidance [18].

## Reference Implementation Plan

- ovm CLI
  - Pull/push via ORAS; local OCI image layout cache [6][7].
  - Inspect/validate manifest against JSON Schema; verify signatures with Cosign [9].
  - Optional: `ovm run` delegating to a runtime daemon (HTTP over Unix socket).

- Runtime adapters
  - Firecracker: map kernel/initrd/virtio‑disks, snapshot restore if provided; fast‑path boot [12].
  - QEMU: qcow2 layering/backing files; IVSHMEM device mapping [10][11].

- Media type registry
  - Publish provisional `vnd.ovms.*` media types in spec repo; track upstreaming to registries and OCI communities [3].

## Conformance Strategy

Conformance areas and tests:
- Schema conformance: JSON Schema validation; negative tests for missing/invalid fields.
- Media types & OCI behavior: ensure artifacts push/pull with correct `artifactType`, annotations, and Referrers discovery [3][18][19].
- Runtime contract: harness that exercises `/start`, `/status`, `/snapshot`, `/stop` on adapters and asserts observable behavior.
- Signing: Cosign verification pathways (success/failure cases) [9].

Outputs:
- Test runner (Go) + fixtures; GitHub Actions matrix over adapters.
- Conformance badges (per runtime) and published results.

## Roadmap and Milestones

Phase 0 — Hardening the draft (2–3 weeks)
- Align spec text with OCI v1.1 (artifactType/subject/referrers); finalize media types [18][19].
- Nail down runtime API and error model; produce OpenAPI schema.
- Produce examples and golden artifacts; wire Cosign verification [9].

Phase 1 — v0.1 RC (4–6 weeks)
- Implement Firecracker and QEMU adapters (MVP paths) [10][11][12].
- ovm CLI MVP (pull, inspect, validate, run).
- Conformance suite v1; publish results for adapters.

Phase 2 — v0.1 Release (2 weeks)
- Spec editorial pass; tag media types; publish docs and examples.
- Outreach to OCI/ORAS/CNCF communities for feedback and alignment.

Phase 3 — v0.2 Planning (subsequent)
- UEFI/multi‑arch matrices; zstd RAM snapshots; richer device catalog.
- Snapshot metadata interoperability and uniqueness handling (VmGenId/MADV_WIPEONSUSPEND) [12].

## Developer Experience

Pushing with ORAS (custom media types)
```bash
oras push registry.example.com/ovms/ubuntu:24.04 \
  --artifact-type application/vnd.ovms.manifest.v1+json \
  manifest.json:application/vnd.ovms.manifest.v1+json \
  kernel.bin:application/vnd.ovms.kernel.v1 \
  initrd.img:application/vnd.ovms.initrd.v1 \
  base.qcow2:application/vnd.oci.image.layer.v1.tar \
  diff1.qcow2:application/vnd.oci.image.layer.v1.tar
```

Discovering related artifacts (signatures/snapshots) via Referrers
```bash
# client support varies; conceptually
curl -sSL \
  https://registry.example.com/v2/ovms/ubuntu/referrers/sha256:<digest>
```

Signing with Cosign
```bash
cosign sign registry.example.com/ovms/ubuntu:24.04
cosign verify registry.example.com/ovms/ubuntu:24.04
```

## Risks and Mitigations

- Divergence across registries for v1.1 features (artifactType/referrers).
  - Mitigate by graceful fallback to annotations and legacy typing; document capabilities [18][19].
- Snapshot interoperability/portability across runtimes.
  - Start with per‑runtime opaque snapshot blob and minimal metadata; converge later [12].
- Security posture of snapshots (key/secret handling).
  - Publish guidance on uniqueness restoration and secret hygiene (VmGenId/MADV_WIPEONSUSPEND) [12].

## Conclusion

OVMS can achieve broad interoperability by standardizing an OCI‑aligned artifact format and a minimal runtime control contract, with a strong conformance and security story. The repository already contains the essential building blocks; this plan sequences the work to deliver a credible v0.1 with real adapters and tests, setting the stage for v0.2 feature growth.

## Sources

1. OCI Distribution Spec — specs.opencontainers.org [https://specs.opencontainers.org/distribution-spec/?v=v1.0.0] [1]
2. OCI Distribution Spec (GitHub) [https://github.com/opencontainers/distribution-spec] [2]
3. OCI Image Spec — Media Types [https://specs.opencontainers.org/image-spec/media-types/] [3]
4. OCI Image Spec — Image Layout [https://specs.opencontainers.org/image-spec/image-layout/] [4]
5. OCI Image Spec v1.1 (PDF) [https://oci-playground.github.io/specs-latest/specs/image/v1.1.0-rc3/oci-image-spec.pdf] [5]
6. ORAS — OCI Artifact Concepts [https://oras.land/docs/concepts/artifact] [6]
7. ORAS — Pushing and Pulling Artifacts [https://oras.land/docs/how_to_guides/pushing_and_pulling/] [7]
8. ORAS — Manifest Config [https://oras.land/docs/how_to_guides/manifest_config/] [8]
9. Sigstore Cosign — Signing Other Types [https://docs.sigstore.dev/cosign/signing/other_types/] [9]
10. QEMU — QCOW2 Format [https://qemu.org/docs/master/interop/qcow2.html] [10]
11. QEMU — IVSHMEM Spec [https://www.qemu.org/docs/master/specs/ivshmem-spec.html] [11]
12. Restoring Uniqueness in MicroVM Snapshots (arXiv) [https://arxiv.org/pdf/2102.12892.pdf] [12]
13. Marc Brooker — Lambda SnapStart and snapshots [https://brooker.co.za/blog/2022/11/29/snapstart.html] [13]
14. WebAssembly Core Spec 2.0 (W3C) [https://www.w3.org/TR/wasm-core-2/] [14]
15. Java SE Specifications (JVM) [https://docs.oracle.com/javase/specs/] [15]
16. ECMA‑335 CLI (6th ed.) [https://www.ecma-international.org/wp-content/uploads/ECMA-335_6th_edition_june_2012.pdf] [16]
17. Ethereum Yellow Paper (PDF) [https://ethereum.github.io/yellowpaper/paper.pdf] [17]
18. Chainguard — Upcoming OCI 1.1 changes (artifactType, subject, referrers) [https://www.chainguard.dev/unchained/oci-announces-upcoming-changes-for-registries] [18]
19. ORAS Artifacts — Manifest Referrers API [https://github.com/oras-project/artifacts-spec/blob/main/manifest-referrers-api.md] [19]
