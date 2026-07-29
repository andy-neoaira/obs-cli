#!/usr/bin/env bash

set -euo pipefail

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

forbidden_patterns=(
  'obs-cli/v'2
  'v'2'.0.0-rc.1'
  'config-v'2'.json'
  'Agent-first V'2
  'V'2'Config'
  'NewV'2'Config'
  'ValidateV'2'Config'
  'renderV'2
  'newV'2'Namespace'
  'new[A-Za-z]+V'2'Command'
  'v'2'Args'
  'daily_notes_v'2
  'link_inspection_v'2
  'metadata_v'2
  'note_operations_v'2
  'search_v'2
  '-v'2'\.schema\.json'
)

scan_content() {
  local status=0
  local pattern
  for pattern in "${forbidden_patterns[@]}"; do
    if rg -n --color never --regexp "$pattern" "$@"; then
      printf 'naming-check: forbidden pattern matched: %s\n' "$pattern" >&2
      status=1
    fi
  done
  return "$status"
}

scan_filenames() {
  local status=0
  local path
  while IFS= read -r path; do
    printf 'naming-check: forbidden versioned filename: %s\n' "$path" >&2
    status=1
  done < <(find "$@" -type f -iname '*v2*' -print)
  return "$status"
}

self_test() {
  local fixture
  fixture=$(mktemp -d "${TMPDIR:-/tmp}/obs-cli-naming-check.XXXXXX")
  trap 'rm -rf "$fixture"' RETURN

  printf '%s\n' 'github.com/gdamore/tcell/v2' >"$fixture/allowed.txt"
  scan_content "$fixture/allowed.txt"

  printf 'protocol=obs-cli/v%s\n' '2' >"$fixture/forbidden.txt"
  if scan_content "$fixture/forbidden.txt" >/dev/null 2>&1; then
    printf 'naming-check self-test: forbidden marker was not detected\n' >&2
    return 1
  fi
  printf 'naming-check self-test: pass\n'
}

if [[ "${1:-}" == "--self-test" ]]; then
  self_test
  exit
fi

cd "$repo_root"

content_targets=(
  cmd
  pkg
  scripts
  skills
  testdata
  docs/spec
  docs/README.md
  README.md
  README_CN.md
  docs/COMMAND_REFERENCE.md
  docs/JOINT_USAGE.md
  docs/TROUBLESHOOTING_AND_RECOVERY.md
  docs/compatibility.json
  .github
  Makefile
  .goreleaser.yml
)

status=0
scan_content "${content_targets[@]}" || status=1
scan_filenames cmd pkg docs/spec || status=1

if ((status != 0)); then
  exit "$status"
fi

printf 'naming-check: pass\n'
