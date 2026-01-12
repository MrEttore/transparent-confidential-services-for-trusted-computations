# Transparent Confidential Service for Trusted Computations

This repository contains the reference implementation of the TCS (Transparent Confidential Service) framework accompanying the paper _"Transparent Confidential Service for Trusted Computations"_. The implementation demonstrates how client-centric attestation transforms opaque remote attestation mechanisms into transparent, user-verifiable protocols for distributed confidential computing services.

## Implementation Overview

The TCS framework addresses a fundamental **trust asymmetry** in confidential computing systems: service providers can verify the confidentiality state of their infrastructure through remote attestation, but end users, the actual data owners, cannot independently verify the provider's confidentiality claims. While providers may assert that their workloads run in genuine TEEs, users must trust these assertions without cryptographic proof. This implementation provides a working prototype that closes this gap, enabling clients to independently verify that their sessions are bound to genuine TEEs executing authorized workloads through transparent, client-initiated attestation protocols.

The system instantiates the **four principal architectural components** defined in the paper (Section 5):

- **Attested Computation Logic:** Intel TDX-enabled confidential virtual machine (CVM) that executes sensitive workloads and serves as the hardware root of trust.
- **Evidence Provider:** In-TEE Go service that aggregates hardware quotes, workload digests, and infrastructure metadata into verifiable evidence bundles.
- **Verifier Application:** Stateless Go service executing appraisal logic under client control, validating evidence against public baselines and endorsements.
- **Relying Application:** React/TypeScript client application orchestrating the challenge-evidence-verification loop and providing transparent visibility into attestation operations.

These components realize the **four trust minimization mechanisms** (Section 4.3):

1. **Nonce-Bound Evidence (M1):** Embeds client-generated challenges into hardware-signed quotes, preventing replay attacks and ensuring session freshness.
2. **Workload Verification Extension (M2):** Enriches attestation with cryptographic workload digests validated against public container registries, extending hardware attestation to application-level integrity.
3. **Verifiable Attestation Bundle (M3):** Enables independent third-party verification through public artifacts and stateless appraisal logic.
4. **User-Facing Attestation Flow (M4):** Exposes the complete attestation protocol in the client UI, providing transparency and progressive disclosure of trust decisions.

## Repository Layout

- `artifacts/`: Reference attestation evidence bundles and Evidence Verifier results that correspond to the framework’s artifact definitions in the paper.
- `evaluation/`: Complete evaluation materials reproducing the paper's Section 6 results, including quantitative performance measurements (K6 load tests, Prometheus metrics) and qualitative security validation (negative control methodology).
- `evidence-provider/`: Go service deployed inside the CVM. Talks to Intel TDX, the container runtime, and cloud metadata services to assemble evidence bundles. Includes a Dockerfile and an opinionated `deploy-and-run.sh` script for pushing binaries to a remote CVM via `gcloud`.
- `evidence-verifier/`: Go service that implements the stateless Verifier Application. Provides HTTP endpoints for quote, workload, and infrastructure appraisal, returning structured attestation results.
- `infrastructure/`: Terraform configuration and bootstrap scripts that provision the attested CVM on Google Cloud with Intel TDX support. The `init-tee.sh` helper prepares Docker, Go, and a dummy workload service inside the CVM.
- `middleware/`: Minimal exemplary middleware component (Nginx-based load balancer) representing the untrusted layer that client challenges and attestation evidence must traverse. Demonstrates sticky-session strategies for routing attestation traffic in distributed deployments.
- `reference-value-provider/`: Repository-backed reference manifests that stand in for the external baseline registry used by the Evidence Verifier.
- `relying-application/`: React SPA. The `features/attestation` feature renders the user-facing attestation flow, issues fresh challenges, and presents evidence and verifier verdicts.

## Implemented Framework Roles

### Attested Computation Logic ([`infrastructure/`](infrastructure/))

Terraform modules provision an Intel TDX-enabled CVM on Google Cloud, configure firewall rules for attestation traffic, and deploy the workload bootstrap (`init-tee.sh`). The VM launches containerized workloads under systemd, ensuring runtime measurements align with published baseline manifests.

### Evidence Provider ([`evidence-provider/`](evidence-provider/))

Built in Go, the Evidence Provider runs inside the Intel TDX CVM alongside its confidential workloads. It exposes REST endpoints that accept base64-encoded challenges, acquire hardware quotes via `go-tdx-guest`, query Docker for container digests and metadata, and capture infrastructure provenance from Google Cloud metadata services. All evidence types are bound to the caller's nonce before being returned as structured JSON, enabling freshness guarantees.

The service generates a self-signed TLS certificate at startup and includes its SHA-256 fingerprint in the attestation quote's `reportData` field, implementing **TLS-bound attestation** (M1). This cryptographically binds the attested hardware identity to the TLS session: `reportData = SHA256(challenge || tlsCertificateFingerprint)`, allowing the Evidence Verifier to link the TLS handshake to the attestation evidence.

Detailed endpoint documentation and API specifications live in [`evidence-provider/README.md`](evidence-provider/README.md).

### Verifier Application ([`evidence-verifier/`](evidence-verifier/))

The Evidence Verifier is a stateless Go service designed to execute in a client-controlled or auditor-controlled environment. Each endpoint parses submitted evidence, fetches public reference values (Intel endorsements, Docker digests, baseline manifests), and emits structured appraisal results. The Evidence Verifier validates **TLS-bound attestation** by checking that the quote's `reportData` correctly encodes `SHA256(challenge || tlsCertificateFingerprint)`, ensuring the attested CVM matches the TLS endpoint. Because no hidden state is maintained, third parties can independently reproduce verification verdicts using only public artifacts, satisfying the verifiability requirement (R3).

Endpoint-level request/response specifications and verification logic details are in [`evidence-verifier/README.md`](evidence-verifier/README.md).

### Relying Application (`relying-application/src/features/attestation`)

The React-based relying party component executes entirely in the user’s client. It generates fresh 64-byte challenges, calls the Evidence Provider, forwards the evidence to the Evidence Verifier, and renders both raw artifacts and human-readable verdicts. Subcomponents such as `AttestationTimeline`, `CloudInfrastructureOverview`, and `IndependentVerificationResources` make the protocol transparent, echoing the user-centric design goals from the paper. A UI-focused walkthrough with screenshot placeholders is available in [`relying-application/README.md`](relying-application/README.md).

For demonstration we host the UI ourselves so that every build originates from a trusted repository, preventing a malicious middleware operator from injecting obfuscated front-end logic that could falsify verification status. In production, the client should control the UI hosting environment to preserve the framework’s trust boundary.

### Reference Value Provider (`reference-value-provider/`)

The Reference Value Provider maintains the public, tamper-evident baselines that anchor the Evidence Verifier’s trust decisions. In production, these references should live in independent registries (for example Sigstore for manifests and Docker Hub for signed container images) so that clients can audit them without involving the service operator. For demonstration, this repository embeds a minimal provider in `reference-value-provider/`, publishing a signed `baseline-manifest.jsonc` that captures the golden CVM image, TDX launch measurements (`mrTd`, `mrSeam`, `teeTcbSvn`, `tdAttributes`, `xfam`), and provenance metadata such as the source image URI and ID. The Evidence Verifier resolves workload digests and infrastructure claims against these artifacts, ensuring that the attested CVM and containers match the public, immutable baselines described in the paper’s model.

Additional implementation notes live in [`reference-value-provider/README.md`](reference-value-provider/README.md).

## Running the Prototype

### Prerequisites

- Go 1.24 (or newer) and Docker for building the Go services locally.
- Node.js 20+ and npm (or pnpm) for the Relying Application.
- Terraform 1.2+ and Google Cloud SDK (`gcloud`) with access to Intel TDX VMs.
- An Intel TDX-enabled machine image available to your Google Cloud project.

### 1. Provision the Computational Logic Attester

```bash
cd infrastructure
cp terraform.tfvars.example terraform.tfvars

# Edit terraform.tfvars with your project, region, and image IDs

gcloud auth application-default login
gcloud config set project <your-project-id>

terraform init
terraform plan
terraform apply
```

The apply step creates the Intel TDX-enabled CVM, installs Docker and Go, and configures a dummy workload service (`llm-core`) under systemd using `init-tee.sh`. Record the VM name, zone, and SSH credentials, these values are required for Evidence Provider deployment.

### 2. Deploy or Run the Evidence Provider

Local development (without hardware-backed quotes) can use `go run`:

```bash
cd evidence-provider
go run cmd/evidenceprovider/main.go
```

For deployment into the CVM and access to genuine Intel TDX quotes:

```bash
cd evidence-provider
cp .deploy.env.example .deploy.env

# Populate REMOTE_USER, REMOTE_HOST, REMOTE_ZONE to match the CVM

go mod tidy

./deploy-and-run.sh
```

The script cross-compiles the binary for Linux/amd64, copies it to the CVM via `gcloud`, and restarts the service on port `8080`. Evidence endpoints become available at `https://<load-balancer-or-vm>:8080`.

### 3. Run the Verifier Application

The Evidence Verifier can run locally or on an independent container:

```bash
cd evidence-verifier
go mod tidy
go run cmd/evidenceverifier/main.go
```

By default the service listens on port `8081`. Use the supplied Dockerfile to containerize the Evidence Verifier when deploying to a production environment.

### 4. Launch the Relying Application

```bash
cd relying-application
npm install

cp .env.example .env
# Update VITE_ATTESTER_URL and VITE_VERIFIER_URL to point at your deployments

npm run dev
```

The Vite development server (default `http://localhost:5173`) renders the attestation experience described in the paper. Update the `.env` entries to point at your deployments, or provide the variables through another configuration mechanism so that the SPA talks to the running Evidence Provider and Verifier instances.

### 5. Optional Supporting Components

- **Exemplary middleware component ([`middleware/`](middleware/)):** Minimal untrusted middleware implementation showing how client challenges traverse intermediary components (load balancers, API gateways) before reaching CVMs. Demonstrates sticky-session affinity patterns for distributed attestation.
- **Performance evaluation ([`evaluation/quantitative/`](evaluation/quantitative/)):** K6 scripts and Prometheus configuration reproducing the performance measurements in Section 6.1.
- **Security validation ([`evaluation/qualitative/`](evaluation/qualitative/)):** Negative control methodology and threat mitigation documentation corresponding to Section 6.2.
- **Attestation artifacts ([`artifacts/`](artifacts/)):** Pre-captured evidence bundles and verification results including positive paths and failure cases for replay attacks, TEE emulation, workload substitution, and evidence tampering.

## Connecting Back to the Paper

Each directory in this repository maps directly to the system roles and mechanisms introduced in the paper’s implementation section:

- **Nonce-bound attestation** is realized by the Evidence Provider’s challenge-binding handlers and the Relying Application’s challenge generator.
- **TLS-bound attestation** extends nonce binding by cryptographically linking the hardware quote to the TLS certificate fingerprint, ensuring the attested CVM identity matches the TLS endpoint.
- **Verifiable evidence bundles** materialize as the JSON payloads returned by the Evidence Provider and validated by the Evidence Verifier against public registries (Intel PCKs, Docker image digests, GitHub-hosted manifests).
- **Workload integrity validation** is enforced through the workload endpoints, ensuring the runtime matches published digests.
- **Transparent user interaction** is implemented in the React feature layer, which exposes every protocol step, underlying artifact, and independent replay resources.

## Testing and Development

### Local Testing with Mock Mode

For rapid development and integration testing without access to Intel TDX hardware, the repository includes a mock mode that simulates hardware attestation primitives. This allows you to test the complete attestation flow locally using Docker Compose.

**Prerequisites:**

- Docker and Docker Compose
- curl (for making HTTP requests)
- jq (recommended, for parsing JSON)

#### Quick Start

Start all services in mock mode:

```bash
docker-compose up --build
```

The Evidence Provider runs on port 8080 (HTTP) and 8443 (HTTPS), the Evidence Verifier on port 8081, and the Relying Application on port 3000.

#### Testing Flow

Test the TLS-bound attestation flow step by step:

```bash
# 1. Generate a challenge (64-byte base64 nonce)
export CHALLENGE_B64="MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMw=="

# 2. Request a TDX quote from the Evidence Provider
RESPONSE=$(curl -k -s -X POST https://localhost:8443/evidence/tdx-quote \
  -H "Content-Type: application/json" \
  -d "{\"challenge\": \"$CHALLENGE_B64\"}")

# 3. Extract the quote and TLS certificate fingerprint
export QUOTE=$(echo $RESPONSE | jq -r '.data.quote')
export FINGERPRINT=$(echo $RESPONSE | jq -r '.data.tlsCertificateFingerprint')

# 4. Verify the quote with the Evidence Verifier
curl -s -X POST http://localhost:8081/verify/tdx-quote \
  -H "Content-Type: application/json" \
  -d "{
    \"issuedChallenge\": \"$CHALLENGE_B64\",
    \"baselineManifestUrl\": \"http://example.com/manifest\",
    \"tlsCertificateFingerprint\": \"$FINGERPRINT\",
    \"quote\": $QUOTE
  }" | jq .
```

Expected result in mock mode:

```json
{
  "status": "success",
  "data": {
    "isVerified": true,
    "message": "tdx quote verified (mock mode - hardware verification skipped)"
  }
}
```

Mock mode uses simulated TDX quotes and skips cryptographic verification of Intel signatures, making it suitable for development workflows. For production validation, deploy to actual Intel TDX-enabled hardware following the instructions in the "Running the Prototype" section above.
