# OVMS Conformance Plan (v0.1)

This document outlines the initial conformance areas and test strategy for OVMS v0.1.

## Areas

1) Schema Conformance
- Validate manifests against `spec/schemas/manifest.v1.json`.
- Negative tests for missing required fields, invalid enums, and mediaType mismatches.

2) OCI Behavior & Media Types
- Push/pull artifacts with `artifactType=application/vnd.ovms.manifest.v1+json`.
- Ensure disk layers carry `application/vnd.oci.image.layer.v1.tar` and `ovms.format` annotation.
- Referrers discovery for related artifacts (signatures, SBOMs, snapshots) using `subject`.

3) Signing & Policy
- Cosign verification pass/fail cases (valid signature, wrong key, unsigned) and policy enforcement.

4) Runtime Contract
- Exercise `spec/api/ovms-runtime.openapi.yaml` endpoints across adapters:
  - start → running within SLO
  - status → states transitions
  - snapshot → publishes artifact
  - stop → graceful termination

## Outputs
- Test runner (Go) executing schema tests and API probes.
- JUnit-style results for CI.
- Conformance badges per adapter runtime.

## Roadmap
- v1: Local tests with a sample OVMS artifact and mock runtime.
- v2: Matrix over Firecracker and QEMU adapters, referrers tests.
