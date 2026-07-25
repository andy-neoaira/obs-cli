#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
matrix="$repo_root/docs/compatibility.json"
skills="$repo_root/skills/evals/scenarios.json"

command -v jq >/dev/null 2>&1 || {
  printf 'compatibility-check: jq is required\n' >&2
  exit 1
}

jq -e '
  .schema_version == "obsidian-joint-compatibility/v1" and
  .content_format_migration_required == false and
  .products["obs-cli"].protocol == "obs-cli/v2" and
  .products["obs-cli"].vault_contract == "vault-contract/v1" and
  .products["miniobsidian.nvim"].cli_optional == true and
  (.compatible_sets | length) >= 1 and
  ([.adapter_modes[].mode] | sort) == ["compatible", "incompatible", "standalone"] and
  (.release_gates.manual_pending | length) >= 1
' "$matrix" >/dev/null

matrix_minimum=$(jq -r '.products.skills.minimum_cli_version' "$matrix")
skills_minimum=$(jq -r '.minimum_cli_version' "$skills")
[ "$matrix_minimum" = "$skills_minimum" ] || {
  printf 'compatibility-check: Skill minimum mismatch (%s != %s)\n' \
    "$matrix_minimum" "$skills_minimum" >&2
  exit 1
}

matrix_skills=$(jq -r '.products.skills.scenario_skills' "$matrix")
actual_skills=$(jq -r '.skills | length' "$skills")
[ "$matrix_skills" = "$actual_skills" ] || {
  printf 'compatibility-check: Skill count mismatch (%s != %s)\n' \
    "$matrix_skills" "$actual_skills" >&2
  exit 1
}

printf 'compatibility-check: matrix, adapter modes, and Skill baseline are consistent\n'
