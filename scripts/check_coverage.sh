#!/usr/bin/env bash
set -euo pipefail

min_coverage="${1:-45.0}"

profile_file="$(mktemp -t attio-cover-XXXXXX.out)"
clean_profile_file="$(mktemp -t attio-cover-clean-XXXXXX.out)"

cleanup() {
  rm -f "$profile_file" "$clean_profile_file"
}
trap cleanup EXIT

go test ./... -coverprofile="$profile_file" >/dev/null

# Some toolchains can duplicate the "mode:" line; keep only the first.
awk 'NR==1 || $1!="mode:"' "$profile_file" > "$clean_profile_file"

total_coverage="$(go tool cover -func="$clean_profile_file" | awk '/^total:/ {gsub(/%/, "", $3); print $3}')"

if [[ -z "$total_coverage" ]]; then
  echo "Failed to parse total coverage" >&2
  exit 1
fi

echo "Total coverage: ${total_coverage}% (minimum: ${min_coverage}%)"

if ! awk -v got="$total_coverage" -v min="$min_coverage" 'BEGIN { exit !(got+0 >= min+0) }'; then
  echo "Coverage gate failed: ${total_coverage}% < ${min_coverage}%" >&2
  exit 1
fi
