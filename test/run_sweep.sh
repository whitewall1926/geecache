#!/usr/bin/env bash
set -euo pipefail

# Auto run concurrent load tests and summarize core metrics.
TARGET="${TARGET:-http://localhost:8002/_geecache/}"
GROUP="${GROUP:-scores}"
MODE="${MODE:-mixed}"
HOT_KEY="${HOT_KEY:-Tom}"
HOT_RATIO="${HOT_RATIO:-80}"
KEYS="${KEYS:-Tom,Jack,Sam}"
N="${N:-50000}"
CONCURRENCIES="${CONCURRENCIES:-100 200 400 800}"
OUT_FILE="${OUT_FILE:-sweep_$(date +%Y%m%d_%H%M%S).tsv}"

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo -e "concurrency\ttotal\tsuccess\tfailed\tqps\tp50\tp95\tp99" > "$OUT_FILE"

echo "Start sweep..."
echo "target=$TARGET mode=$MODE group=$GROUP n=$N concurrencies=[$CONCURRENCIES]"

for c in $CONCURRENCIES; do
  echo "\n== Run concurrency: $c =="
  output=$(go run client.go \
    -target "$TARGET" \
    -group "$GROUP" \
    -mode "$MODE" \
    -hot-key "$HOT_KEY" \
    -hot-ratio "$HOT_RATIO" \
    -keys "$KEYS" \
    -n "$N" \
    -c "$c")

  echo "$output"

  total=$(echo "$output" | awk -F': *' '/^Total Requests:/ {print $2}')
  success=$(echo "$output" | awk -F': *' '/^Success/ {print $2}')
  failed=$(echo "$output" | awk -F': *' '/^Failed/ {print $2}')
  qps=$(echo "$output" | awk -F': *' '/^QPS/ {print $2}')
  p50=$(echo "$output" | awk -F': *' '/^P50 Latency/ {print $2}')
  p95=$(echo "$output" | awk -F': *' '/^P95 Latency/ {print $2}')
  p99=$(echo "$output" | awk -F': *' '/^P99 Latency/ {print $2}')

  echo -e "$c\t$total\t$success\t$failed\t$qps\t$p50\t$p95\t$p99" >> "$OUT_FILE"
done

echo "\nSweep finished. Summary saved to: $SCRIPT_DIR/$OUT_FILE"
cat "$OUT_FILE"
