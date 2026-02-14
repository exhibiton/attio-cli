#!/usr/bin/env bash
set -euo pipefail

profile_file="${1:-coverage.out}"

if [[ ! -f "$profile_file" ]]; then
  echo "coverage profile not found: $profile_file" >&2
  exit 1
fi

clean_profile_file="$(mktemp -t attio-cover-clean-XXXXXX.out)"
trap 'rm -f "$clean_profile_file"' EXIT

awk 'NR==1 || $1!="mode:"' "$profile_file" > "$clean_profile_file"
go tool cover -func="$clean_profile_file"
