#!/usr/bin/env bash

set -euo pipefail

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_root"

retired_markers=(
  'COMPAT_CLEANUP_TASKS\.md'
  'SKILL_SCENARIOS\.md'
  'PROJECT_REVIEW\.md'
  'V1_RC_VALIDATION\.md'
  'docs/agent-first-v2'
  'docs/migration'
  'docs/v1-clean-start-refactor'
)

scan_targets=(
  README.md
  README_CN.md
  docs
  skills
  scripts
  .github
)

status=0
for marker in "${retired_markers[@]}"; do
  if rg -n --color never --glob '!scripts/docs-check.sh' --regexp "$marker" "${scan_targets[@]}"; then
    printf 'docs-check: retired document reference matched: %s\n' "$marker" >&2
    status=1
  fi
done

while IFS= read -r file; do
  while IFS= read -r raw; do
    target=${raw#']('}
    target=${target%')'}
    target=${target#'<'}
    target=${target%'>'}
    case "$target" in
      ''|'#'*|http://*|https://*|mailto:*|app://*)
        continue
        ;;
    esac
    target=${target%%#*}
    if [[ ! -e "$(dirname "$file")/$target" ]]; then
      printf 'docs-check: broken local link: %s -> %s\n' "$file" "$target" >&2
      status=1
    fi
  done < <(grep -oE '\]\([^)]*\)' "$file" || true)
done < <(
  find docs skills -type f -name '*.md' -print
  printf '%s\n' README.md README_CN.md THIRD_PARTY_NOTICES.md
)

if ((status != 0)); then
  exit "$status"
fi

printf 'docs-check: current documentation references are valid\n'
