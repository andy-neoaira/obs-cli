#!/usr/bin/env bash

set -euo pipefail

source_dir="/Users/andy/github/obs-cli/testdata/vault-contract"
target_dir="/Users/andy/github/miniobsidian.nvim/tests/fixtures/vault-contract"
mode="${1:---sync}"

if [[ ! -d "$source_dir" ]]; then
  echo "canonical fixture directory not found: $source_dir" >&2
  exit 1
fi

case "$mode" in
  --sync)
    mkdir -p "$target_dir"
    rsync -a --delete "$source_dir/" "$target_dir/"
    ;;
  --check)
    if [[ ! -d "$target_dir" ]]; then
      echo "fixture copy not found: $target_dir" >&2
      exit 1
    fi
    diff -qr "$source_dir" "$target_dir"
    ;;
  *)
    echo "usage: $0 [--sync|--check]" >&2
    exit 2
    ;;
esac
