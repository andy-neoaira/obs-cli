#!/usr/bin/env bash

set -euo pipefail

version="${RC_VERSION:-${SKILL_EVAL_CLI_VERSION:-v1.0.0-rc.1}}"
output_dir="$(mktemp -d "${TMPDIR:-/tmp}/obs-cli-rc.XXXXXX")"
trap 'rm -rf "$output_dir"' EXIT
binary="$output_dir/obs-cli"
candidate_binary="${RC_BINARY:-}"

grep -Fq 'cmd.ldflagsVersion={{.Tag}}' .goreleaser.yml || {
  echo "RC smoke failed: GoReleaser must inject the full tag via {{.Tag}}" >&2
  exit 1
}

if [ -n "$candidate_binary" ]; then
  [ -x "$candidate_binary" ] || {
    echo "RC smoke failed: RC_BINARY is not executable: $candidate_binary" >&2
    exit 1
  }
  binary="$candidate_binary"
else
  GOMODCACHE="$output_dir/modcache" go build -mod=vendor \
    -ldflags "-s -w -X github.com/andy-neoaira/obs-cli/cmd.ldflagsVersion=$version" \
    -o "$binary" .
fi

"$binary" capabilities --output json --require note.get --require note.patch |
  jq -e --arg version "$version" \
    '
      .ok == true and
      .protocol_version == "obs-cli/v1" and
      .data.cli_version == $version and
      .data.protocol_versions == ["obs-cli/v1"] and
      .data.vault_contract.implemented == "vault-contract/v1" and
      (.data.operations | length > 0) and
      all(.data.feature_flags | keys[]; endswith("_v" + "2") | not)
    ' \
    >/dev/null

"$binary" --version | grep -Fqx "obs-cli version $version"

GOMODCACHE="$output_dir/modcache" go test ./cmd -run TestRCSmoke -count=1
echo "V1 RC smoke passed for $version"
