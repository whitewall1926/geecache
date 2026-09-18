#!/usr/bin/env bash

set -euo pipefail

minimum_coverage=80
coverage_file="$(mktemp)"
trap 'rm -f "$coverage_file"' EXIT

go test ./... -coverprofile="$coverage_file"
coverage="$(go tool cover -func="$coverage_file" | awk '$1 == "total:" { gsub("%", "", $3); print $3 }')"

if [[ -z "$coverage" ]]; then
    echo "Unable to determine total test coverage" >&2
    exit 1
fi

if ! awk -v coverage="$coverage" -v minimum="$minimum_coverage" 'BEGIN { exit !(coverage >= minimum) }'; then
    echo "Test coverage ${coverage}% is below the ${minimum_coverage}% minimum" >&2
    exit 1
fi

echo "Total test coverage: ${coverage}% (minimum: ${minimum_coverage}%)"
