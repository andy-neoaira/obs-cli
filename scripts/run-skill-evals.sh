#!/bin/sh

set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_root"

jq empty skills/evals/scenarios.json skills/evals/scenarios.schema.json
./scripts/lint-skills.sh --strict
./scripts/install-evals.sh
go test ./cmd -run '^TestSkillEval' -count=1
