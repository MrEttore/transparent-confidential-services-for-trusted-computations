# Transparent Confidential Services for Trusted Computations

This repository contains the reference implementation that accompanies the paper _Transparent Confidential Services for Trusted Computations_. The code demonstrates how client-centric attestation can be applied to distributed confidential AI workloads by exposing verifiable evidence to end users. Readers can inspect the exact services, infrastructure manifests, and user interfaces that realize the framework described in the paper.

The implementation instantiates the four principal roles introduced in the paper:

- **Computational Logic Attester:** a confidential VM that runs the sensitive workload and exposes attestable state.
- **Evidence Provider:** an in-TEE service that aggregates hardware quotes, workload digests, and infrastructure metadata.
- **Verifier Application:** an independent appraisal service that validates attestation evidence against public baselines and endorsements.
- **Relying Application:** a client-side feature that orchestrates the challenge-evidence-verification loop for end users.

Together, these components operationalize the four attestation mechanisms highlighted in the paper: the _User-Facing Attestation Flow_, the _Verifiable Attestation Bundle_, the _Workload Verification Extension_, and the _Nonce-Bound Evidence_.

## Repository Layout

- `artifacts/` - Reference attestation evidence bundles and verifier results that correspond to the framework’s artifact definitions in the paper.
- `evidence-provider/` - Go service deployed inside the confidential VM. Talks to Intel TDX, the container runtime, and cloud metadata services to assemble evidence bundles. Includes a Dockerfile and an opinionated `deploy-and-run.sh` script for pushing binaries to a remote CVM via `gcloud`.
- `evidence-verifier/` - Go service that implements the stateless Verifier Application. Provides HTTP endpoints for quote, workload, and infrastructure appraisal, returning structured attestation results.
- `infrastructure/` - Terraform configuration and bootstrap scripts that provision the Computational Logic Attester on Google Cloud with Intel TDX support. The `init-tee.sh` helper prepares Docker, Go, and the workload service inside the CVM.
- `load-tests/` - K6 scripts and payloads used to characterize attestation latency under load.
- `middleware/` - Nginx and Flask demo showing sticky-session strategies that keep multi-round attestation conversations anchored to a single CVM instance.
- `reference-value-provider/` - Repository-backed reference manifests that stand in for the external baseline registry used by the verifier.
- `relying-application/` - React single-page application. The `features/attestation` feature renders the user-facing attestation flow, issues fresh challenges, and presents evidence and verifier verdicts.

## Implemented Framework Roles

### Computational Logic Attester (`infrastructure/`)

Terraform modules create a confidential VM on Google Cloud, configure firewall ingress for attestation traffic, and install the workload bootstrap (`init-tee.sh`). The VM launches the production workload container (`llm-core`) under systemd, ensuring the runtime measurements align with the published baseline manifests. This matches the paper’s description of a CVM acting as the root of trust for confidential AI inference.

### Evidence Provider (`evidence-provider/`)

Built in Go, the Evidence Provider runs alongside the confidential workload inside the Intel TDX-based CVM. It exposes REST endpoints that accept a base64 challenge, acquire a hardware quote via `go-tdx-guest`, query Docker for the live container digest, and capture provenance from Google Cloud metadata. All evidence types are bound to the caller’s nonce before being returned as JSON, enabling the freshness guarantees discussed in the paper. Detailed endpoint documentation lives in [`evidence-provider/README.md`](evidence-provider/README.md).

### Verifier Application (`evidence-verifier/`)

The Verifier is a stateless Go service intended to execute in a client-controlled or auditor-controlled environment. Each endpoint parses the submitted evidence, fetches public reference values (Intel endorsement certificates, workload digests, infrastructure manifests), and emits an appraisal result. Because no hidden state is kept server-side, third parties can reproduce verdicts, satisfying the paper’s verifiability requirement. Endpoint-level request/response details are captured in [`evidence-verifier/README.md`](evidence-verifier/README.md).

### Relying Application (`relying-application/src/features/attestation`)

The React-based relying party component executes entirely in the user’s client. It generates fresh 64-byte challenges, calls the Evidence Provider, forwards the evidence to the Evidence Verifier, and renders both raw artifacts and human-readable verdicts. Subcomponents such as `AttestationTimeline`, `CloudInfrastructureOverview`, and `IndependentVerificationResources` make the protocol transparent, echoing the user-centric design goals from the paper. A UI-focused walkthrough with screenshot placeholders is available in [`relying-application/README.md`](relying-application/README.md).

For demonstration we host the UI ourselves so that every build originates from a trusted repository, preventing a malicious middleware operator from injecting obfuscated front-end logic that could falsify verification status. In production, the client (or auditor) should control the UI hosting environment to preserve the framework’s trust boundary.

### Reference Value Provider (`reference-value-provider/`)

The Reference Value Provider maintains the public, tamper-evident baselines that anchor the verifier’s trust decisions. In the deployed framework, these references live in independent registries (for example Sigstore for manifests and Docker Hub for signed container images) so that clients can audit them without involving the service operator. For demonstration, this repository embeds a minimal provider in `reference-value-provider/`, publishing a signed `baseline-manifest.jsonc` that captures the golden CVM image, TDX launch measurements (`mrTd`, `mrSeam`, `teeTcbSvn`, `tdAttributes`, `xfam`), and provenance metadata such as the source image URI and ID. The Evidence Verifier resolves workload digests and infrastructure claims against these artifacts, ensuring that the attested CVM and containers match the public, immutable baselines described in the paper’s model.

Additional implementation notes live in [`reference-value-provider/README.md`](reference-value-provider/README.md).

## Running the Prototype

Run the orchestration commands from a macOS or Linux workstation (the scripts assume a POSIX shell and native tooling). The confidential workloads themselves execute remotely on Google Cloud Intel TDX VMs. Replace placeholder values with your environment specifics when following the steps below.

### Prerequisites

- Go 1.24 (or newer) and Docker for building the Go services locally.
- Node.js 20+ and npm (or pnpm) for the relying application.
- Terraform 1.2+ and Google Cloud SDK (`gcloud`) with access to Intel TDX-enabled projects.
- An Intel TDX-enabled machine image available to your Google Cloud project (see `infrastructure/terraform.tfvars.example`).

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

The apply step creates the confidential VM, installs Docker and Go, and deploys the `llm-core` workload under systemd using `init-tee.sh`. Record the VM name, zone, and SSH username—these values are required by the Evidence Provider deployment script.

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

The Verifier can run locally or on an independent VM/container:

```bash
cd evidence-verifier
go mod tidy
go run cmd/evidenceverifier/main.go
```

By default the service listens on port `8081` and exposes `/verify/tdx-quote`, `/verify/workloads`, and `/verify/infrastructure` endpoints (plus `/` for health). Use the supplied Dockerfile to containerize the verifier when deploying to a production environment.

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

- **Middleware sticky-session demo (`middleware/`):** Illustrates the load-balancer affinity strategies needed to keep the challenge–response handshake attached to a single CVM. See the folder README for IP-hash versus cookie-based setups.
- **Load testing (`load-tests/`):** K6 scripts reproduce the performance evaluation referenced in the paper. Adjust target URLs before running `k6 run ...`.
- **Attestation artifacts (`artifacts/`):** Pre-captured JSON evidence and verification results align with the figures and discussion in the paper’s implementation and evaluation sections.

## Connecting Back to the Paper

Each directory in this repository maps directly to the system roles and mechanisms introduced in the paper’s implementation section:

- **Nonce-bound attestation** is realized by the Evidence Provider’s challenge-binding handlers and the Relying Application’s challenge generator.
- **Verifiable evidence bundles** materialize as the JSON payloads returned by the Evidence Provider and validated by the Evidence Verifier against public registries (Intel PCKs, Docker image digests, GitHub-hosted manifests).
- **Workload integrity validation** is enforced through the workload endpoints and Terraform-managed golden images, ensuring the runtime matches published digests.
- **Transparent user interaction** is implemented in the React feature layer, which exposes every protocol step, underlying artifact, and independent replay resources.
