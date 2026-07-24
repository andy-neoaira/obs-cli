#!/usr/bin/env bash

set -euo pipefail

required_files=(
  "LICENSE"
  "THIRD_PARTY_NOTICES.md"
  "vendor/modules.txt"
)

for file in "${required_files[@]}"; do
  if [[ ! -s "$file" ]]; then
    echo "required license file is missing or empty: $file" >&2
    exit 1
  fi
done

grep -Fq "Copyright (c) 2023 Kartikay Jainwal" LICENSE
grep -Fq "Copyright (c) 2026 andy-neoaira" LICENSE
grep -Fq "Yakitrak/notesmd-cli" README.md
grep -Fq "Yakitrak/notesmd-cli" README_CN.md
grep -Fq "Yakitrak/notesmd-cli" THIRD_PARTY_NOTICES.md

missing=0
while read -r module; do
  module_dir="vendor/$module"
  if ! find "$module_dir" -maxdepth 1 -type f \
    \( -iname 'LICENSE*' -o -iname 'COPYING*' \) -print -quit | grep -q .; then
    echo "vendored module has no license file: $module" >&2
    missing=1
  fi
  if ! grep -Fq "\`$module\`" THIRD_PARTY_NOTICES.md; then
    echo "vendored module missing from THIRD_PARTY_NOTICES.md: $module" >&2
    missing=1
  fi
done < <(awk '/^# [^#]/{print $2}' vendor/modules.txt)

if [[ "$missing" -ne 0 ]]; then
  exit 1
fi

grep -Fq -- "- LICENSE" .goreleaser.yml
grep -Fq -- "- THIRD_PARTY_NOTICES.md" .goreleaser.yml
grep -Fq -- "- vendor/**/LICENSE*" .goreleaser.yml
grep -Fq -- "- vendor/**/NOTICE" .goreleaser.yml

echo "license check passed"
