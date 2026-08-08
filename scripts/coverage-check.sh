#!/usr/bin/env bash

set -euo pipefail

threshold="${COVERAGE_MIN:-70.0}"
profile="$(mktemp "${TMPDIR:-/tmp}/obs-cli-coverage.XXXXXX")"
list_module_cache="$(mktemp -d "${TMPDIR:-/tmp}/obs-cli-coverage-modcache.XXXXXX")"
trap 'rm -f "$profile"; rm -rf "$list_module_cache"' EXIT

test_output=$(go test ./... -coverprofile="$profile")
printf '%s\n' "$test_output"
total="$(go tool cover -func="$profile" | awk '/^total:/{gsub("%", "", $3); print $3}')"
awk -v actual="$total" -v minimum="$threshold" 'BEGIN {
  if (actual + 0 < minimum + 0) {
    printf "coverage %.1f%% is below %.1f%%\n", actual, minimum > "/dev/stderr"
    exit 1
  }
  printf "coverage %.1f%% meets %.1f%% threshold\n", actual, minimum
}'

check_package() {
  local package=$1
  local minimum=$2
  local actual
  actual=$(awk -v package="$package" '
    $2 == package {
      for (i = 1; i <= NF; i++) {
        if ($i == "coverage:") {
          value = $(i + 1)
          gsub("%", "", value)
          print value
        }
      }
    }
  ' <<<"$test_output")
  [[ -n "$actual" ]] || {
    printf 'coverage for critical package %s was not reported\n' "$package" >&2
    return 1
  }
  awk -v package="$package" -v actual="$actual" -v minimum="$minimum" 'BEGIN {
    if (actual + 0 < minimum + 0) {
      printf "%s coverage %.1f%% is below %.1f%%\n", package, actual, minimum > "/dev/stderr"
      exit 1
    }
    printf "%s coverage %.1f%% meets %.1f%% threshold\n", package, actual, minimum
  }'
}

package_threshold="${PACKAGE_COVERAGE_MIN:-80.1}"
packages=$(GOMODCACHE="$list_module_cache" go list -mod=vendor ./...)
while IFS= read -r package; do
  check_package "$package" "$package_threshold"
done <<<"$packages"
