# Evaluation

This folder contains all materials used to evaluate the TCS framework presented in this repository.

Our evaluation is structured along two complementary dimensions:

1. **Quantitative (performance)**: measures the performance overhead introduced by the attestation mechanisms.
2. **Qualitative (security validation / negative controls)**: validates security-relevant behavior through systematic negative controls and documents expected failure modes.

The goal is to make the experiments reproducible and to keep both the scripts/configuration and raw artifacts/results in one place.

## Where to Start

### Quantitative evaluation ([quantitative/](quantitative/))

See the [README](quantitative/README.md) for: how to run load tests, required tooling, Prometheus configuration, scripts, and where results are stored.

### Qualitative evaluation ([qualitative/](qualitative/))

See the [README](qualitative/README.md) for: the negative-control methodology, how to reproduce checks, and how artifacts are organized.
