 # OVMS Signing Guide (Cosign + OCI Referrers)

 This guide shows how to sign OVMS artifacts using Sigstore Cosign and how to verify signatures using the OCI Referrers graph (OCI 1.1: `artifactType`, `subject`).

 ## Prerequisites
 - Cosign installed (`cosign version`)
 - ORAS or crane for registry interactions
 - Access to an OCI registry

 ## Media Types and Relationships
 - OVMS manifest artifactType: `application/vnd.ovms.manifest.v1+json`
 - Signatures are separate artifacts that set `subject` to the OVMS manifest digest and are discoverable via the Referrers API.

 ## 1) Generate Keys (optional; you can also use keyless OIDC)
 ```bash
 cosign generate-key-pair
 # produces cosign.key and cosign.pub
 ```

 ## 2) Push OVMS Artifact
 Use ORAS to push the OVMS manifest + blobs with proper media types.
 ```bash
 oras push $REG/repo/ubuntu:24.04 ./ovm \
   --media-type application/vnd.ovms.manifest.v1+json \
   --artifact-type application/vnd.ovms.manifest.v1+json
 ```

 Record the digest (example):
 ```bash
 DIGEST=$(oras discover -o json $REG/repo/ubuntu:24.04 | jq -r '.manifests[0].digest')
 ```

 ## 3) Sign by Digest
 ```bash
 cosign sign --key cosign.key $REG/repo@${DIGEST}
 # keyless: COSIGN_EXPERIMENTAL=1 cosign sign $REG/repo@${DIGEST}
 ```

 ## 4) Verify Signature
 ```bash
 cosign verify --key cosign.pub $REG/repo@${DIGEST}
 # keyless: COSIGN_EXPERIMENTAL=1 cosign verify $REG/repo@${DIGEST}
 ```

 ## 5) Discover Referrers (Signatures, SBOMs, Snapshots)
 ```bash
 oras discover --artifact-type application/vnd.dev.cosign.artifact.sig $REG/repo@${DIGEST}
 ```

 ## Policy Suggestions
 - Admission controllers (or Huo plugin) SHOULD require valid signatures for OVMS manifests before runtime start.
 - Maintain a trusted keys/policies set for verification.

 ## Troubleshooting
 - Ensure registry supports Referrers API; otherwise use fallback (oci-layout or oras annotations).
 - If using Docker Hub, consider ORAS 1.1+ and enable OCI artifacts support where applicable.