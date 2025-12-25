#!/bin/bash
set -euo pipefail

CLIENT_ID="client-unknown"
SCRAPE_WAIT=20

PROM_URL="http://127.0.0.1:9090"
EV_URL="http://127.0.0.1:8081"

RESET_PROM=true
RESET_EV=false

usage() {
  echo "Usage: $0 --client <client-sm|client-md|client-lg> [--scrape-wait <sec>] [--reset-prom] [--no-reset-prom] [--reset-ev]"
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --client) CLIENT_ID="${2:-client-unknown}"; shift 2 ;;
    --scrape-wait) SCRAPE_WAIT="${2:-20}"; shift 2 ;;
    --reset-prom) RESET_PROM=true; shift ;;
    --no-reset-prom) RESET_PROM=false; shift ;;
    --reset-ev) RESET_EV=true; shift ;;
    *) usage ;;
  esac
done

RUN_ID="$(date -u +%Y%m%dT%H%M%SZ)-${CLIENT_ID}"
RESULTS_DIR="$HOME/results/$RUN_ID"
mkdir -p "$RESULTS_DIR"/{k6,prometheus,notes}

echo "$RUN_ID" | tee "$RESULTS_DIR/notes/run_id.txt" >/dev/null
echo "RESULTS_DIR=$RESULTS_DIR" | tee "$RESULTS_DIR/notes/results_dir.txt" >/dev/null

EP_DIR="$HOME/quantitative/evidence-provider"
EV_DIR="$HOME/quantitative/evidence-verifier"

command -v k6 >/dev/null 2>&1 || { echo "ERROR: k6 not found. Run prepare-load-tests.sh first."; exit 1; }
test -f "$HOME/prometheus.yml" || { echo "ERROR: ~/prometheus.yml missing. Run prepare-load-tests.sh first."; exit 1; }

wait_for_prometheus() {
  echo "==> Waiting for Prometheus readiness at $PROM_URL/-/ready"
  for i in $(seq 1 60); do
    if curl -4 -fsS "$PROM_URL/-/ready" >/dev/null 2>&1; then
      echo "==> Prometheus is ready."
      return 0
    fi
    if ! sudo docker ps --format '{{.Names}}' | grep -qx prometheus; then
      echo "ERROR: Prometheus container is not running."
      sudo docker ps -a --filter name=prometheus || true
      sudo docker logs --tail=200 prometheus || true
      return 1
    fi
    sleep 1
  done
  echo "ERROR: Prometheus did not become ready within 60s."
  sudo docker logs --tail=200 prometheus || true
  return 1
}

if [[ "$RESET_EV" == "true" ]]; then
  if sudo docker ps -a --format '{{.Names}}' | grep -qx evidenceverifier; then
    sudo docker rm -f evidenceverifier >/dev/null 2>&1 || true
  fi
  sudo docker run --rm -d -p 8081:8081 --name evidenceverifier attestify/evidenceverifier:latest >/dev/null
fi

if [[ "$RESET_PROM" == "true" ]]; then
  if sudo docker ps -a --format '{{.Names}}' | grep -qx prometheus; then
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
fi

wait_for_prometheus

curl -4 -fsS "$EV_URL/metrics" >/dev/null 2>&1 || {
  echo "ERROR: Evidence Verifier not reachable on $EV_URL/metrics"
  sudo docker ps || true
  sudo docker logs --tail=200 evidenceverifier || true
  exit 1
}

{
  echo "UTC: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "Client: $CLIENT_ID"
  echo "Run ID: $RUN_ID"
  echo "Hostname: $(hostname)"
  echo
  echo "uname -a:"; uname -a
  echo
  echo "lscpu:"; lscpu || true
  echo
  echo "free -h:"; free -h || true
  echo
  echo "ip -br addr:"; ip -br addr || true
} > "$RESULTS_DIR/notes/vm_info.txt"

{
  echo "k6: $(k6 version || true)"
  echo "docker: $(docker --version 2>/dev/null || true)"
  echo "prometheus image: $(sudo docker inspect --format '{{.Config.Image}}' prometheus 2>/dev/null || echo 'n/a')"
  echo "evidenceverifier image: $(sudo docker inspect --format '{{.Config.Image}}' evidenceverifier 2>/dev/null || echo 'n/a')"
} > "$RESULTS_DIR/notes/versions.txt"

TIMES_CSV="$RESULTS_DIR/notes/test_times.csv"
echo "test_id,service,script,start_epoch,end_epoch,start_iso,end_iso" > "$TIMES_CSV"

run_one() {
  local test_id="$1" service="$2" script_path="$3" out_json="$4"
  [[ -f "$script_path" ]] || { echo "ERROR: missing script: $script_path"; exit 1; }

  echo "==> Running ${test_id} (${service}): ${script_path}"
  local start_epoch start_iso end_epoch end_iso
  start_epoch="$(date -u +%s)"
  start_iso="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

  k6 run \
    --tag "run_id=${RUN_ID}" \
    --tag "client=${CLIENT_ID}" \
    --tag "test_id=${test_id}" \
    --tag "service=${service}" \
    --summary-export "$out_json" \
    "$script_path"

  end_epoch="$(date -u +%s)"
  end_iso="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "${test_id},${service},${script_path},${start_epoch},${end_epoch},${start_iso},${end_iso}" >> "$TIMES_CSV"

  echo "Sleeping ${SCRAPE_WAIT}s to allow Prometheus scrape after test..."
  sleep "$SCRAPE_WAIT"
}

run_one "ep-tdx-quote"      "ep" "$EP_DIR/k6-evidence-quote.js"     "$RESULTS_DIR/k6/ep-tdx-quote.json"
run_one "ep-workload"       "ep" "$EP_DIR/k6-evidence-workload.js"  "$RESULTS_DIR/k6/ep-workload.json"
run_one "ep-infrastructure" "ep" "$EP_DIR/k6-evidence-infra.js"     "$RESULTS_DIR/k6/ep-infrastructure.json"

run_one "ev-tdx-quote"      "ev" "$EV_DIR/k6-verify-tdx.js"         "$RESULTS_DIR/k6/ev-tdx-quote.json"
run_one "ev-workloads"      "ev" "$EV_DIR/k6-verify-workloads.js"   "$RESULTS_DIR/k6/ev-workloads.json"
run_one "ev-infrastructure" "ev" "$EV_DIR/k6-verify-infra.js"       "$RESULTS_DIR/k6/ev-infrastructure.json"

echo
echo "Load suite complete."
echo "Run ID:     $RUN_ID"
echo "Results in: $RESULTS_DIR"
echo
echo "NEXT (manual):"
echo "  export RESULTS_DIR=\"$RESULTS_DIR\""
echo "  ~/export-prometheus-results.sh"
