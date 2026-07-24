#!/usr/bin/env bash

set -euo pipefail

output_dir="$(mktemp -d "${TMPDIR:-/tmp}/obs-cli-build.XXXXXX")"
trap 'rm -rf "$output_dir"' EXIT

for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64; do
  os="${target%/*}"
  arch="${target#*/}"
  suffix=""
  if [[ "$os" == "windows" ]]; then
    suffix=".exe"
  fi
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 GOMODCACHE="$output_dir/modcache" \
    go build -mod=vendor -o "$output_dir/obs-cli_${os}_${arch}${suffix}" .
done

echo "cross-platform build check passed"
