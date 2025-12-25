#!/bin/bash
set -euo pipefail

REPO_OWNER="MrEttore"
REPO_NAME="transparent-confidential-services-for-trusted-computations"
REPO_REF="main"
RAW_BASE="https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/${REPO_REF}"

WORKDIR="$HOME/attestation-client"
EVIDENCE_VERIFIER_IMAGE="attestify/evidenceverifier:latest"
BASE_URL_MIDDLEWARE="https://middleware-40815901860.europe-west4.run.app"
BASE_URL_EVIDENCE_VERIFIER="http://127.0.0.1:8081"

mkdir -p "$WORKDIR"

echo "==> Installing prerequisites..."
for pkg in docker.io docker-doc docker-compose docker-compose-v2 podman-docker containerd runc; do
  sudo apt-get remove -y "$pkg" >/dev/null 2>&1 || true
done

sudo apt-get update -y
sudo apt-get install -y ca-certificates curl gnupg lsof jq

sudo install -m 0755 -d /etc/apt/keyrings
if [ ! -f /etc/apt/keyrings/docker.asc ]; then
  sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
  sudo chmod a+r /etc/apt/keyrings/docker.asc
fi

CODENAME="$(. /etc/os-release && echo "${VERSION_CODENAME}")"
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu ${CODENAME} stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null

sudo apt-get update -y
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

sudo systemctl enable --now docker
sudo systemctl status docker --no-pager >/dev/null

echo "==> Starting Evidence Verifier container fresh..."
sudo docker pull "$EVIDENCE_VERIFIER_IMAGE" >/dev/null

if sudo docker ps -a --format '{{.Names}}' | grep -qx evidenceverifier; then
  sudo docker rm -f evidenceverifier >/dev/null 2>&1 || true
fi

sudo docker run --rm -d \
  -p 8081:8081 \
  --name evidenceverifier \
  "$EVIDENCE_VERIFIER_IMAGE" >/dev/null

dl() { curl -fL --retry 3 --retry-delay 1 -o "$2" "$1"; }

echo "==> Downloading manual payloads..."
dl "${RAW_BASE}/evaluation/quantitative/evidence-provider/test-payloads/challenge.json" \
   "$WORKDIR/challenge.json"
dl "${RAW_BASE}/evaluation/quantitative/evidence-verifier/test-payloads/quote.json" \
   "$WORKDIR/tdx-quote.json"
dl "${RAW_BASE}/evaluation/quantitative/evidence-verifier/test-payloads/workloads.json" \
   "$WORKDIR/workloads.json"
dl "${RAW_BASE}/evaluation/quantitative/evidence-verifier/test-payloads/infra.json" \
   "$WORKDIR/infrastructure.json"

jq . < "$WORKDIR/challenge.json" >/dev/null
jq . < "$WORKDIR/tdx-quote.json" >/dev/null
jq . < "$WORKDIR/workloads.json" >/dev/null
jq . < "$WORKDIR/infrastructure.json" >/dev/null

cat > "$WORKDIR/README-manual-tests.txt" <<EOT
Manual testing instructions
===========================

Hardcoded targets:
- Evidence Provider (via middleware): ${BASE_URL_MIDDLEWARE}
- Evidence Verifier (local container): ${BASE_URL_EVIDENCE_VERIFIER}

1) Evidence Provider (via middleware):
   curl -sS -X POST "${BASE_URL_MIDDLEWARE}/evidence/tdx-quote" \\
     -H "Content-Type: application/json" \\
     --data-binary @"$WORKDIR/challenge.json" | jq .

2) Evidence Verifier (local container):
   curl -sS -X POST "${BASE_URL_EVIDENCE_VERIFIER}/verify/tdx-quote" \\
     -H "Content-Type: application/json" \\
     --data-binary @"$WORKDIR/tdx-quote.json" | jq .

   curl -sS -X POST "${BASE_URL_EVIDENCE_VERIFIER}/verify/workloads" \\
     -H "Content-Type: application/json" \\
     --data-binary @"$WORKDIR/workloads.json" | jq .

   curl -sS -X POST "${BASE_URL_EVIDENCE_VERIFIER}/verify/infrastructure" \\
     -H "Content-Type: application/json" \\
     --data-binary @"$WORKDIR/infrastructure.json" | jq .
EOT

echo
echo "==> Done."
echo "Workdir: $WORKDIR"
echo "Verifier container: evidenceverifier (listening on :8081)"
echo "Manual instructions: $WORKDIR/README-manual-tests.txt"
