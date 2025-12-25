#!/bin/bash
set -euo pipefail

: "${RESULTS_DIR:?Set RESULTS_DIR first (e.g., export RESULTS_DIR=~/results/<RUN_ID>)}"

PROM_URL="http://127.0.0.1:9090"
PROM_CONTAINER="prometheus"

EP_METRIC="ep_http_handler_seconds"
EV_METRIC="ev_http_handler_seconds"

RATE_WIN="5m"
STEP="5s"

OUT_DIR="$RESULTS_DIR/prometheus"
API_DIR="$OUT_DIR/api"
SNAP_DIR="$OUT_DIR/tsdb_snapshot"
PROMQL_DIR="$OUT_DIR/promql"

mkdir -p "$API_DIR" "$SNAP_DIR" "$PROMQL_DIR/instant" "$PROMQL_DIR/range"

curl -4 -fsS "$PROM_URL/-/ready" >/dev/null 2>&1 || {
  echo "ERROR: Prometheus not reachable at $PROM_URL"
  exit 1
}

sudo docker ps --format '{{.Names}}' | grep -qx "$PROM_CONTAINER" || {
  echo "ERROR: container '$PROM_CONTAINER' not running"
  exit 1
}

sudo apt-get update -y >/dev/null
sudo apt-get install -y jq >/dev/null

echo "==> Exporting Prometheus API artifacts..."
curl -sS "$PROM_URL/api/v1/targets"            | jq . > "$API_DIR/targets.json"
curl -sS "$PROM_URL/api/v1/status/config"      | jq . > "$API_DIR/config.json"
curl -sS "$PROM_URL/api/v1/status/runtimeinfo" | jq . > "$API_DIR/runtimeinfo.json"
curl -sS "$PROM_URL/api/v1/status/buildinfo"   | jq . > "$API_DIR/buildinfo.json"
curl -sS "$PROM_URL/api/v1/status/flags"       | jq . > "$API_DIR/flags.json"

TSDB_ABS="$(jq -r '.data["storage.tsdb.path"]' "$API_DIR/flags.json")"
TSDB_ABS="${TSDB_ABS%/}"

echo "==> Creating TSDB snapshot via Admin API..."
SNAP_RESP="$(curl -sS -XPOST "$PROM_URL/api/v1/admin/tsdb/snapshot")"
echo "$SNAP_RESP" | jq . > "$API_DIR/tsdb_snapshot_response.json"

SNAP_NAME="$(echo "$SNAP_RESP" | jq -r '.data.name')"
SNAP_ABS="${TSDB_ABS}/snapshots/${SNAP_NAME}"

echo "==> Snapshot created: $SNAP_NAME"
echo "==> Expecting snapshot at: $SNAP_ABS"
echo "==> Waiting for snapshot directory to appear..."

for _ in $(seq 1 60); do
  if sudo docker exec "$PROM_CONTAINER" sh -c "test -d '$SNAP_ABS'"; then
    break
  fi
  sleep 1
done

sudo docker exec "$PROM_CONTAINER" sh -c "test -d '$SNAP_ABS'" || {
  echo "ERROR: snapshot directory did not appear: $SNAP_ABS"
  sudo docker exec "$PROM_CONTAINER" sh -c "ls -la '$TSDB_ABS' || true; ls -la '$TSDB_ABS/snapshots' || true" || true
  exit 1
}

echo "==> Copying snapshot from container..."
sudo docker cp "$PROM_CONTAINER:$SNAP_ABS" "$SNAP_DIR/$SNAP_NAME"

TIMES_CSV="$RESULTS_DIR/notes/test_times.csv"
[[ -f "$TIMES_CSV" ]] || { echo "ERROR: missing $TIMES_CSV"; exit 1; }

query_range() {
  local q="$1" start="$2" end="$3" step="$4" out="$5"
  curl -sG "$PROM_URL/api/v1/query_range" \
    --data-urlencode "query=$q" \
    --data-urlencode "start=$start" \
    --data-urlencode "end=$end" \
    --data-urlencode "step=$step" \
    | jq . > "$out"
}

query_instant() {
  local q="$1" ts="$2" out="$3"
  curl -sG "$PROM_URL/api/v1/query" \
    --data-urlencode "query=$q" \
    --data-urlencode "time=$ts" \
    | jq . > "$out"
}

echo "==> Exporting PromQL results from test windows: $TIMES_CSV"

tail -n +2 "$TIMES_CSV" | while IFS=',' read -r test_id service script start_epoch end_epoch start_iso end_iso; do
  start="$start_epoch"
  end="$((end_epoch + 30))"

  window_sec="$((end - start))"
  WINDOW="${window_sec}s"

  if [[ "$service" == "ep" ]]; then
    case "$test_id" in
      ep-tdx-quote)      route="evidence_tdx_quote" ;;
      ep-workload)       route="evidence_workloads" ;;
      ep-infrastructure) route="evidence_infrastructure" ;;
      *) route="" ;;
    esac
    METRIC="$EP_METRIC"
  else
    case "$test_id" in
      ev-tdx-quote)      route="verify_tdx_quote" ;;
      ev-workloads)      route="verify_workloads" ;;
      ev-infrastructure) route="verify_infrastructure" ;;
      *) route="" ;;
    esac
    METRIC="$EV_METRIC"
  fi

  [[ -n "$route" ]] || { echo "WARN: unmapped test_id=$test_id"; continue; }

  echo "  -> $test_id ($service) $start_iso .. $end_iso  route=$route  window=$WINDOW"

  Q_P50_RATE="histogram_quantile(0.50, sum by (le) (rate(${METRIC}_bucket{route=\"${route}\"}[${RATE_WIN}]))) * 1000"
  Q_P95_RATE="histogram_quantile(0.95, sum by (le) (rate(${METRIC}_bucket{route=\"${route}\"}[${RATE_WIN}]))) * 1000"
  Q_AVG_RATE="(sum(rate(${METRIC}_sum{route=\"${route}\"}[${RATE_WIN}])) / sum(rate(${METRIC}_count{route=\"${route}\"}[${RATE_WIN}]))) * 1000"
  Q_RPS_RATE="sum(rate(${METRIC}_count{route=\"${route}\"}[${RATE_WIN}]))"

  Q_P50_WIN="histogram_quantile(0.50, sum by (le) (increase(${METRIC}_bucket{route=\"${route}\"}[${WINDOW}]))) * 1000"
  Q_P95_WIN="histogram_quantile(0.95, sum by (le) (increase(${METRIC}_bucket{route=\"${route}\"}[${WINDOW}]))) * 1000"
  Q_AVG_WIN="(sum(increase(${METRIC}_sum{route=\"${route}\"}[${WINDOW}])) / sum(increase(${METRIC}_count{route=\"${route}\"}[${WINDOW}]))) * 1000"
  Q_TOTAL_WIN="sum(increase(${METRIC}_count{route=\"${route}\"}[${WINDOW}]))"

  query_range "$Q_P50_RATE" "$start" "$end" "$STEP" "$PROMQL_DIR/range/${test_id}-p50-rate.json"
  query_range "$Q_P95_RATE" "$start" "$end" "$STEP" "$PROMQL_DIR/range/${test_id}-p95-rate.json"
  query_range "$Q_AVG_RATE" "$start" "$end" "$STEP" "$PROMQL_DIR/range/${test_id}-avg-rate.json"
  query_range "$Q_RPS_RATE" "$start" "$end" "$STEP" "$PROMQL_DIR/range/${test_id}-rps-rate.json"

  query_instant "$Q_P50_WIN"  "$end" "$PROMQL_DIR/instant/${test_id}-p50-window.json"
  query_instant "$Q_P95_WIN"  "$end" "$PROMQL_DIR/instant/${test_id}-p95-window.json"
  query_instant "$Q_AVG_WIN"  "$end" "$PROMQL_DIR/instant/${test_id}-avg-window.json"
  query_instant "$Q_TOTAL_WIN" "$end" "$PROMQL_DIR/instant/${test_id}-total-window.json"
done

echo
echo "Prometheus export complete. Written under: $OUT_DIR"
echo "Next (manual):"
echo "  tar -C \"$HOME/results\" -czf \"$HOME/results/$(basename "$RESULTS_DIR").tar.gz\" \"$(basename "$RESULTS_DIR")\""
