# Attestation Artifacts

This directory aggregates the canonical payloads referenced in the paper to illustrate the attestation protocol flow. The JSON files capture both sides of the protocol: evidence emitted by the in-CVM Evidence Provider and the appraisal results returned by the Evidence Verifier. Hosting these payloads alongside the codebase gives readers a concrete view of the artifacts described in the _Attestation Evidence_ and _Attestation Results_ sections of the paper.

## Purpose

- Document the exact field structure of each evidence type (attestation quote, workload metadata, infrastructure metadata) as produced by the prototype.
- Demonstrate how the Evidence Verifier expresses pass and failure conditions for the three evidence appraisal endpoints.
- Provide reproducible fixtures that developers can replay when extending or auditing the framework.

## Directory Overview

- `attestation-evidence/`: Evidence payloads returned by the in-CVM Evidence Provider service.
- `attestation-result/`: Result payloads emitted by the Evidence Verifier endpoints for the same session inputs.

## Evidence Bundles (`attestation-evidence/`)

The Evidence Provider exposes dedicated endpoints for each artifact, and the corresponding JSON responses are mirrored here:

- `attestation-quote-evidence.json`: Hardware-backed Intel TDX quote that binds the session nonce (`reportData`) to the launch measurements (`mrTd`, `mrSeam`, `rtmrs`, `tdAttributes`, `teeTcbSvn`, `xfam`).
- `workload-evidence.json`: Container inventory with digests, runtime state, and provenance labels collected directly from the Docker engine for workload integrity checks.
- `infrastructure-metadata-evidence.json`: Cloud context describing the CVM (instance type, region, shielding status) and boot disk lineage so the Evidence Verifier can trace the machine back to the published baseline manifest.

> These payloads map 1:1 to the _Attestation Evidence_ presented in the paper and can be replayed against the Evidence Verifier for regression testing.

## Verification Results (`attestation-result/`)

Each subdirectory captures responses from the Evidence Verifier's REST endpoints and includes both success and representative failure cases:

- `quote-attestation-result/`: Validates quote authenticity, endorsement chains, and nonce freshness. Contains `pass.json` plus failure scenarios such as `fail-mrTd-mismatch.json` and `fail-nonce-mismatch.json`.
- `workload-attestation-result/`: Confirms container digests against the Reference Value Provider, illustrating how provenance mismatches are surfaced.
- `infrastructure-metadata-attestation-result/`: Checks CVM configuration and boot disk metadata against the baseline manifest, demonstrating how deviations are reported.

> The result payloads align with the _Attestation Result_ description in the paper and show exactly what a Relying Application would receive when invoking the Evidence Verifier endpoints.
