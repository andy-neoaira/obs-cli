#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
mini_root=${MINIOBSIDIAN_ROOT:-"$repo_root/../miniobsidian.nvim"}
nvim_bin=${NVIM:-nvim}
plenary_root=${PLENARY_DIR:-"$HOME/.local/share/nvim/lazy/plenary.nvim"}

fail() {
  printf 'three-client-e2e: %s\n' "$1" >&2
  exit 1
}

command -v jq >/dev/null 2>&1 || fail "jq is required"
command -v "$nvim_bin" >/dev/null 2>&1 || fail "Neovim is required: $nvim_bin"
[ -f "$mini_root/tests/three_client_e2e_spec.lua" ] ||
  fail "miniobsidian.nvim E2E spec not found under $mini_root"
[ -f "$plenary_root/lua/plenary/init.lua" ] ||
  fail "plenary.nvim not found under $plenary_root; set PLENARY_DIR"

run_root=$(mktemp -d "${TMPDIR:-/tmp}/obs-cli-three-client.XXXXXX")
case "$run_root" in
  "${TMPDIR:-/tmp}"/obs-cli-three-client.*) ;;
  *) fail "refusing unsafe temporary directory: $run_root" ;;
esac
cleanup() {
  rm -rf -- "$run_root"
}
trap cleanup EXIT HUP INT TERM

vault_root="$run_root/vault"
config_root="$run_root/config"
summary_file="$run_root/summary.json"
actual_normalized="$run_root/actual.normalized.json"
golden_normalized="$run_root/golden.normalized.json"
cli_bin="$run_root/obs-cli"
go_cache="$run_root/go-cache"
go_mod_cache="$run_root/mod-cache"

mkdir -p "$vault_root" "$config_root" "$go_cache" "$go_mod_cache"
cp -R "$repo_root/testdata/three-client/seed/." "$vault_root/"

if [ -n "${OBS_CLI_BIN:-}" ]; then
  [ -x "$OBS_CLI_BIN" ] || fail "OBS_CLI_BIN is not executable: $OBS_CLI_BIN"
  cp "$OBS_CLI_BIN" "$cli_bin"
else
  command -v go >/dev/null 2>&1 || fail "go is required"
  (
    cd "$repo_root"
    GOCACHE="$go_cache" GOMODCACHE="$go_mod_cache" go build -mod=vendor -o "$cli_bin" .
  )
fi

export OBS_CLI_CONFIG_HOME="$config_root"
"$cli_bin" vault add "$vault_root" --name ThreeClientE2E --set-default --output json |
  jq -e '.ok == true and .operation == "vault.add"' >/dev/null

export THREE_CLIENT_E2E=1
export THREE_CLIENT_VAULT="$vault_root"
export THREE_CLIENT_CLI="$cli_bin"
export THREE_CLIENT_WORKDIR="$run_root"
export THREE_CLIENT_SUMMARY="$summary_file"
export MINIOBSIDIAN_ROOT="$mini_root"
export PLENARY_DIR="$plenary_root"
export NVIM_LOG_FILE="$run_root/nvim.log"

(
  cd "$mini_root"
  "$nvim_bin" --headless -u tests/minimal_init.lua \
    -c "PlenaryBustedDirectory tests/three_client_e2e_spec.lua { minimal_init = 'tests/minimal_init.lua' }" \
    -c qa
)

[ -f "$summary_file" ] || fail "Neovim E2E did not write a summary"
jq -S . "$summary_file" >"$actual_normalized"
jq -S . "$repo_root/testdata/three-client/golden-summary.json" >"$golden_normalized"
cmp -s "$actual_normalized" "$golden_normalized" ||
  fail "actual summary differs from testdata/three-client/golden-summary.json"

printf 'three-client-e2e: 6/6 automated scenarios passed\n'
