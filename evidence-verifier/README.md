# Evidence Verifier Service

The Evidence Verifier implements the client-controlled appraisal logic. It validates the _Attestation Evidence_ returned by the Evidence Provider against public baselines (Intel endorsements, Docker image digests, infrastructure manifests). This document captures the HTTP interface, request/response payloads, and observed behaviour.

## API Overview

| Method | Path                     | Purpose                                                                                 |
| ------ | ------------------------ | --------------------------------------------------------------------------------------- |
| `POST` | `/verify/tdx-quote`      | Verify a nonce-bound Intel TDX quote against Intel endorsements and baseline manifests. |
| `POST` | `/verify/workloads`      | Verify reported workload containers and images against trusted registry metadata.       |
| `POST` | `/verify/infrastructure` | Verify infrastructure metadata against baseline manifests.                              |
| `GET`  | `/`                      | Liveness/health check.                                                                  |
| `GET`  | `/metrics`               | Prometheus metrics exposition.                                                          |
| `GET`  | `/debug/pprof/*`         | Go pprof handlers (development only).                                                   |

All verification endpoints expect the Relying Application to forward the evidence bundle returned by the Evidence Provider along with the original challenge (for freshness) and any additional reference inputs.

### POST /verify/tdx-quote

**Purpose:** Validate that the supplied Intel TDX quote is authentic, that its `reportData` matches the caller-issued challenge, and that its measurements align with the provided baseline manifest.

**Request:**

```jsonc
{
    // Challenge originally issued by the Relying Application.
    "issuedChallenge": "Qk1nQ2h6b2J4b2d6b2J4Qk1nQ2h6b2J4b2d6b2J4Qk1nQ2h6b2J4b2d6b2J4Qk1nQ2h6b2J4b2d6b2J4Qk1n2g==",

    // HTTPS URL pointing at the golden manifest (JSONC) published by the Reference Value Provider.
    "baselineManifestUrl": "https://anonymous.4open.science/r/transparent-confidential-services-76BB/artifacts/baseline-manifest-v2.jsonc",

    // TDX quote payload obtained from the Evidence Provider.
    "quote": { "header": { ... }, "tdQuoteBody": { ... }, "signedData": { ... } }
}
```

**Response:**

```jsonc
{
  "status": "success",
  "data": {
    "isVerified": true,
    "message": "TDX quote matches baseline and challenge.",
    "timestamp": "2025-06-15T09:25:03.964712Z"
  }
}
```

- `200 OK` when verification succeeds.
- `400 Bad Request` for malformed JSON or missing fields.
- `422 Unprocessable Entity` when verification fails (e.g., challenge mismatch, signature failure).

### POST /verify/workloads

**Purpose:** Validate that the reported workload containers/images originate from the expected registry digest and that the challenge is preserved in the evidence.

**Request:**

```jsonc
{
    "issuedChallenge": "Qk1nQ2h6b2J4b2d6b2J4Qk1nQ2h6b2J4b2d6b2J4Qk1nQ2h6b2J4b2d6b2J4Qk1nQ2h6b2J4b2d6b2J4Qk1n2g==",

    "referenceImage": {
        "namespace": "sanctuairy",
        "repository": "llm-core",
        "tag": "latest"
    },

    "evidence": {
        "containers": [ { "id": "03e3f866...", "imageDigest": "sha256:1fa274d3...", ... } ],
        "images": [ { "id": "sha256:78a811ac...", "repoDigests": ["sanctuairy/llm-core@sha256:..."] } ],
        "reportData": "Qk1nQ2h6b2J4b2d6...=="
    }
}
```

**Response:**

```jsonc
{
  "status": "success",
  "data": {
    "isVerified": true,
    "message": "Workload digest matches sanctuairy/llm-core:latest",
    "goldenImageDigest": "sha256:ef903eae47acc61ad494fcfe5eb51441cb81465e6131f3532bcb9309eb4210e0",
    "providedImageDigest": "sha256:1fa274d318656b6cdf70210930ad8dee5069e6c93765937d8af41013ec0db5f6",
    "referenceImage": "sanctuairy/llm-core:latest",
    "timestamp": "2025-06-15T09:25:05.112349Z"
  }
}
```

- `200 OK` when digests and challenge validation succeed.
- `400 Bad Request` for malformed JSON or missing reference information.
- `422 Unprocessable Entity` when digest comparison or challenge validation fails.

### POST /verify/infrastructure

**Purpose:** Validate that the reported GCP instance metadata lines up with the baseline manifest, including boot disk provenance and confidential computing configuration.

**Request:**

```jsonc
{
    "issuedChallenge": "Qk1nQ2h6b2J4b2d6b2J4Qk1nQ2h6b2J4b2d6b2J4Qk1nQ2h6b2J4b2d6b2J4Qk1nQ2h6b2J4b2d6b2J4Qk1n2g==",

    "baselineManifestUrl": "https://anonymous.4open.science/r/transparent-confidential-services-76BB/artifacts/baseline-manifest.jsonc",

    "evidence": {
        "summary": { "provider": "Google Cloud Platform", "instanceId": "8105972874086498876", ... },
        "instance": { "cpuPlatform": "Intel Sapphire Rapids", ... },
        "disk": { "sourceImage": "https://www.googleapis.com/compute/v1/projects/...", ... },
        "reportData": "Qk1nQ2h6b2J4b2d6...=="
    }
}
```

**Response:**

```jsonc
{
  "status": "success",
  "data": {
    "isVerified": true,
    "message": "Infrastructure evidence matches baseline manifest.",
    "verifiedBootDiskSourceImage": "https://www.googleapis.com/compute/v1/projects/sanctuairy/global/images/golden-reference-tee",
    "verifiedBootDiskSourceImageID": "2293782220127998194",
    "timestamp": "2025-06-15T09:25:06.004582Z"
  }
}
```

- `200 OK` when provenance verification succeeds.
- `400 Bad Request` for malformed JSON or missing manifest URL.
- `422 Unprocessable Entity` when evidence differs from the baseline or the challenge mismatches.

## Error Envelope

On failure the service responds with:

```json
{
  "status": "error",
  "message": "human-readable description"
}
```

HTTP status codes mirror the handler outcome (`400` for request validation errors, `422` for verification failures, `500` for unexpected internal issues).

## Operational Notes

- The service listens on HTTP port `8081`.
- `/metrics` exposes Prometheus request metrics and `/debug/pprof/*` is intended for controlled debugging.
- Keep the baseline manifests and Docker registry credentials (if required) accessible from the environment hosting the Evidence Verifier.
