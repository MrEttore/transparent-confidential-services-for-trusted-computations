# Reference Value Provider

The Reference Value Provider role anchors the Evidence Verifier's trust decisions in public, immutable artifacts. In a production deployment these artifacts would be hosted on independent registries such as Sigstore (for manifests) and Docker Hub (for container images). For the repository-backed implementation of the framework, the `reference-value-provider/` directory stands in for that external source of truth by publishing the signed baseline manifests that the Evidence Verifier queries during appraisal.

## Baseline Manifest

The consolidated manifest lives at `baseline-manifest.jsonc` and records the trusted launch state for the CVM. The JSONC structure contains:

- `version`, `createdAt`, `author`: changelog metadata that is signed via Git tags (for example `git tag -s v1.0.0 ...`).
- `sourceImage`, `sourceImageId`: provenance links to the golden boot disk exposed through Google Compute Engine.
- `mrTd`, `mrSeam`, `tdAttributes`, `teeTcbSvn`, `xfam`: Intel TDX launch measurements extracted from the reference CVM.

The Evidence Verifier compares live _Attestation Evidence_ against these values to confirm that the running CVM replays the same measured environment. Container digests are validated separately against their authoritative registries. The manifest exists solely to cover infrastructure and TEE state.

## Golden Reference Generation Workflow

The baseline manifest was produced once, following the workflow below.

1. **Create the reference CVM:**

   ```bash
   gcloud compute instances create reference-cvm \
     --zone="europe-west4-a" \
     --machine-type="c3-standard-4" \
     --confidential-compute-type="TDX" \
     --maintenance-policy="TERMINATE" \
     --image="ubuntu-2204-jammy-v20250430" \
     --image-project="ubuntu-os-cloud" \
     --boot-disk-size="30GB" \
     --shielded-secure-boot \
     --tags="allow-attestation,allow-service-requests" \
     --scopes https://www.googleapis.com/auth/cloud-platform \
     --metadata-from-file startup-script="init-tee.sh"
   ```

2. **Wait for initialization:** Allow the startup script to install Docker, Go, and enable the confidential workload service.

3. **Deploy the Evidence Provider:** Copy the compiled Evidence Provider binary to the CVM and start it so that attestation APIs are reachable.

4. **Collect launch measurements:** Call the Evidence Provider running on the reference CVM to export the combined evidence bundle (TDX quote, workload metadata, infrastructure metadata). Record the TDX measurements (`mrTd`, `mrSeam`, `teeTcbSvn`, `tdAttributes`, `xfam`).

5. **Build and sign the baseline manifest:**

   ```jsonc
   {
     // Metadata
     "version": "v1.0.0",
     "createdAt": "2025-12-01T12:00:00Z",
     "author": "John.Doe@dummy-distributed-confidential-services.com",

     // Infrastructure-specific fields
     "sourceImage": "https://www.googleapis.com/compute/v1/projects/cvm-icdcs/global/images/golden-reference-cvm",
     "sourceImageId": "1708120587032331792",

     // TDX-specific fields
     "mrTd": "pYROiIl7cMMYvvkp7039bHMExSxLycPzkTLw/czs8+tbq3ARDuQqElCaMcA3KIaU",
     "mrSeam": "v7NgrI5iM6G8oUM8r3OC2VwWW0p3+wC/FDXloI8wDN/q1e5oRhr9m2xyjc51NGAt",
     "tdAttributes": "AAAAEAAAAAA=",
     "teeTcbSvn": "CAEIAAAAAAAAAAAAAAAAAA==",
     "xfam": "5wAGAAAAAAA="
   }
   ```

   ```bash
   git add baseline-manifest.jsonc
   git commit -m "Add baseline manifest for golden-reference-cvm"
   git tag -s v1.0.0 -m "Initial golden baseline for production VM image"
   git push --follow-tags
   ```

6. **Snapshot the boot disk:**

   ```bash
   gcloud compute images create golden-reference-cvm \
     --source-disk="reference-cvm" \
     --source-disk-zone="europe-west4-a"

   gcloud compute images add-iam-policy-binding golden-reference-cvm \
     --member="allAuthenticatedUsers" \
     --role="roles/compute.imageUser"
   ```

7. **Instantiate attesters from the golden image:**

   ```bash
   gcloud compute instances create cvm-1 \
     --zone="europe-west4-a" \
     --machine-type="c3-standard-4" \
     --confidential-compute-type="TDX" \
     --maintenance-policy="TERMINATE" \
     --image="golden-reference-cvm" \
     --image-project="cvm-icdcs" \
     --boot-disk-size="30GB" \
     --shielded-secure-boot \
     --tags="allow-attestation,allow-service-requests" \
     --scopes=https://www.googleapis.com/auth/cloud-platform
   ```

In a production deployment these artifacts should be published to an external registry and protected using a supply-chain hardening framework. The repository copy remains to make the framework self-contained for evaluation and demonstration.
