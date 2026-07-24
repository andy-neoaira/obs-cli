#!/usr/bin/env bash

set -euo pipefail

threshold="${COVERAGE_MIN:-70.0}"
profile="$(mktemp "${TMPDIR:-/tmp}/obs-cli-coverage.XXXXXX")"
trap 'rm -f "$profile"' EXIT

go test ./... -coverprofile="$profile"
total="$(go tool cover -func="$profile" | awk '/^total:/{gsub("%", "", $3); print $3}')"
awk -v actual="$total" -v minimum="$threshold" 'BEGIN {
  if (actual + 0 < minimum + 0) {
    printf "coverage %.1f%% is below %.1f%%\n", actual, minimum > "/dev/stderr"
    exit 1
  }
  printf "coverage %.1f%% meets %.1f%% threshold\n", actual, minimum
}'
