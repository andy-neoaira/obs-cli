#!/usr/bin/env bash

set -euo pipefail

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

legacy_paths=(
  pkg/actions
  pkg/obsidian/fuzzyfinder.go
  pkg/obsidian/note.go
  pkg/obsidian/uri.go
  pkg/obsidian/vault.go
  pkg/obsidian/vault_list.go
  pkg/obsidian/vault_path.go
)

legacy_patterns=(
  'github\.com/andy-neoaira/obs-cli/pkg/actions'
  'github\.com/ktr0731/go-fuzzyfinder'
  'github\.com/skratchdot/open-golang'
  '\bFuzzyFinderManager\b'
  '\bLinkRewriteManager\b'
  '\bNoteManager\b'
  '\bUriManager\b'
  '\bVaultManager\b'
)

check_paths() {
  local root=$1
  local status=0
  local relative

  for relative in "${legacy_paths[@]}"; do
    if [[ -e "$root/$relative" ]]; then
      printf 'architecture-check: legacy runtime path exists: %s\n' "$relative" >&2
      status=1
    fi
  done
  return "$status"
}

check_content() {
  local status=0
  local pattern
  shift

  for pattern in "${legacy_patterns[@]}"; do
    if rg -n --color never --regexp "$pattern" "$@"; then
      printf 'architecture-check: legacy runtime pattern matched: %s\n' "$pattern" >&2
      status=1
    fi
  done
  return "$status"
}

self_test() {
  local fixture
  fixture=$(mktemp -d "${TMPDIR:-/tmp}/obs-cli-architecture-check.XXXXXX")
  trap 'rm -rf "$fixture"' RETURN

  mkdir -p "$fixture/pkg/actions"
  if check_paths "$fixture" >/dev/null 2>&1; then
    printf 'architecture-check self-test: legacy path was not detected\n' >&2
    return 1
  fi

  printf '%s\n' 'type VaultManager interface{}' >"$fixture/legacy.go"
  if check_content ignored "$fixture/legacy.go" >/dev/null 2>&1; then
    printf 'architecture-check self-test: legacy symbol was not detected\n' >&2
    return 1
  fi

  printf '%s\n' 'type Registry struct{}' >"$fixture/current.go"
  check_content ignored "$fixture/current.go"
  printf 'architecture-check self-test: pass\n'
}

if [[ "${1:-}" == "--self-test" ]]; then
  self_test
  exit
fi

cd "$repo_root"

status=0
check_paths "$repo_root" || status=1
check_content ignored cmd pkg go.mod go.sum vendor/modules.txt || status=1

if ((status != 0)); then
  exit "$status"
fi

printf 'architecture-check: pass\n'
