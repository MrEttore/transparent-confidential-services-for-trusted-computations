#!/bin/bash
set -euo pipefail

REPO_OWNER="MrEttore"
REPO_NAME="transparent-confidential-services-for-trusted-computations"
REPO_REF="main"
RAW_BASE="https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/${REPO_REF}"

EP_DIR="$HOME/quantitative/evidence-provider"
EV_DIR="$HOME/quantitative/evidence-verifier"

mkdir -p "$EP_DIR/test-payloads" "$EV_DIR/test-payloads"

echo "==> Installing tools..."
sudo apt-get update -y
sudo apt-get install -y curl tar wget jq

dl() { curl -fL --retry 3 --retry-delay 1 -o "$2" "$1"; }

version_lt() {
  local a="${1#v}" b="${2#v}"
  [[ "$(printf '%s\n%s\n' "$a" "$b" | sort -V | head -n1)" != "$b" ]]
}

resolve_latest_k6_tag() {
  curl -fsSL "https://api.github.com/repos/grafana/k6/releases" \
    | jq -r '[.[] | select(.prerelease==false)][0].tag_name'
}

install_k6_from_tag() {
  local tag="$1"
  local ver="${tag#v}"
  local tmpdir
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  echo "==> Installing k6 ${tag}..."
  local url="https://github.com/grafana/k6/releases/download/${tag}/k6-v${ver}-linux-amd64.tar.gz"
  (cd "$tmpdir" && wget -q "$url" -O k6.tgz && tar -xzf k6.tgz)
  sudo mv "$tmpdir/k6-v${ver}-linux-amd64/k6" /usr/local/bin/k6
  sudo chmod +x /usr/local/bin/k6
}

ensure_latest_k6() {
  local latest_tag
  latest_tag="$(resolve_latest_k6_tag)"

  if [[ -z "$latest_tag" || "$latest_tag" == "null" ]]; then
    echo "ERROR: Could not resolve latest k6 version from GitHub API."
    exit 1
  fi

  local installed_tag=""
  if command -v k6 >/dev/null 2>&1; then
    installed_tag="$(k6 version | awk '{print $2}' | head -n1)"
  fi

  if [[ -z "$installed_tag" ]]; then
    install_k6_from_tag "$latest_tag"
  else
    if version_lt "$installed_tag" "$latest_tag"; then
      echo "==> Upgrading k6 from ${installed_tag} to ${latest_tag}..."
      install_k6_from_tag "$latest_tag"
    else
      echo "==> k6 already up-to-date (${installed_tag})."
    fi
  fi

  echo "==> k6 version: $(k6 version)"
}

ensure_latest_k6

echo "==> Downloading Evidence Provider payload + k6 scripts from repo..."
dl "${RAW_BASE}/evaluation/quantitative/evidence-provider/test-payloads/challenge.json" \
   "$EP_DIR/test-payloads/challenge.json"
for f in k6-evidence-quote.js k6-evidence-workload.js k6-evidence-infra.js; do
  dl "${RAW_BASE}/evaluation/quantitative/evidence-provider/${f}" "$EP_DIR/${f}"
done

echo "==> Downloading Evidence Verifier payloads + k6 scripts from repo..."
for p in infra.json quote.json workloads.json; do
  dl "${RAW_BASE}/evaluation/quantitative/evidence-verifier/test-payloads/${p}" \
     "$EV_DIR/test-payloads/${p}"
done
for f in k6-verify-tdx.js k6-verify-workloads.js k6-verify-infra.js; do
  dl "${RAW_BASE}/evaluation/quantitative/evidence-verifier/${f}" "$EV_DIR/${f}"
done

echo "==> Downloading prometheus.yml from repo..."
dl "${RAW_BASE}/evaluation/quantitative/prometheus.yml" "$HOME/prometheus.yml"

echo "==> Restarting Prometheus container fresh..."
if sudo docker ps -a --format '{{.Names}}' | grep -qx 'prometheus'; then
  sudo docker rm -f prometheus >/dev/null 2>&1 || true
fi

sudo docker run -d \
  --name prometheus \
  --network host \
  -v "$HOME/prometheus.yml:/etc/prometheus/prometheus.yml:ro" \
  prom/prometheus:v2.50.1 \
  --config.file=/etc/prometheus/prometheus.yml \
  --storage.tsdb.path=/prometheus \
  --web.enable-admin-api >/dev/null

echo
echo "==> Done."
echo "Repo ref: $REPO_REF"
echo "EP dir:   $EP_DIR"
echo "EV dir:   $EV_DIR"
echo "Prom UI:  http://localhost:9090"
echo
echo "Check targets:"
echo "  curl -sS http://localhost:9090/api/v1/targets | jq -r '.data.activeTargets[] | \"\\(.labels.job) \\(.labels.instance) \\(.health) \\(.lastError // \"\")\"'"
