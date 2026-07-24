#!/usr/bin/env bash

set -euo pipefail

files=()
while IFS= read -r file; do
  files+=("$file")
done < <(git ls-files '*.go' ':!vendor/**')
unformatted="$(gofmt -l "${files[@]}")"
if [[ -n "$unformatted" ]]; then
  echo "Go files require gofmt:" >&2
  echo "$unformatted" >&2
  exit 1
fi
