#!/usr/bin/env bash

set -euo pipefail

version="${RC_VERSION:-${SKILL_EVAL_CLI_VERSION:-v2.0.0-rc.1}}"
output_dir="$(mktemp -d "${TMPDIR:-/tmp}/obs-cli-rc.XXXXXX")"
trap 'rm -rf "$output_dir"' EXIT
binary="$output_dir/obs-cli"

GOMODCACHE="$output_dir/modcache" go build -mod=vendor \
  -ldflags "-s -w -X github.com/andy-neoaira/obs-cli/cmd.ldflagsVersion=$version" \
  -o "$binary" .

"$binary" capabilities --output json --require note.get --require note.patch |
  jq -e --arg version "$version" \
    '.ok == true and .data.cli_version == $version and (.data.operations | length > 0)' \
    >/dev/null

GOMODCACHE="$output_dir/modcache" go test ./cmd -run TestV2RCSmoke -count=1
echo "V2 RC smoke passed for $version"
