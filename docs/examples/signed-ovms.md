 # Example: Sign and Verify an OVMS Artifact

 This walkthrough signs an OVMS manifest and verifies it via OCI Referrers.

 ## Steps

 1) Build/prepare an OVMS artifact directory `./ovm` with `manifest.json` and blobs.

 2) Push to registry:
 ```bash
 REG=registry.example.com/ovms
 oras push $REG/ubuntu:24.04 ./ovm \
   --media-type application/vnd.ovms.manifest.v1+json \
   --artifact-type application/vnd.ovms.manifest.v1+json
 ```

 3) Get digest and sign:
 ```bash
 DIGEST=$(oras discover -o json $REG/ubuntu:24.04 | jq -r '.manifests[0].digest')
 cosign sign --key cosign.key $REG/ubuntu@${DIGEST}
 ```

 4) Verify signature:
 ```bash
 cosign verify --key cosign.pub $REG/ubuntu@${DIGEST}
 ```

 5) Discover referrers:
 ```bash
 oras discover --artifact-type application/vnd.dev.cosign.artifact.sig $REG/ubuntu@${DIGEST}
 ```

 Expected: Cosign signature object(s) listed referencing the base OVMS manifest.
