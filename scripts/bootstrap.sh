#!/usr/bin/env bash

set -euo pipefail

repository="${OBS_CLI_REPOSITORY:-andy-neoaira/obs-cli}"
agent="codex"
version=""
install_dir="${OBS_CLI_INSTALL_DIR:-${HOME}/.local/bin}"
skills_dir=""
source_root=""
force_cli=false
upgrade_skills=false
dry_run=false

usage() {
  cat <<'EOF'
Install obs-cli and its official Skills as one explicit workflow.

Usage:
  bootstrap.sh [--agent <name>] [--version <tag>] [--install-dir <dir>]
               [--skills-dir <dir>] [--source <checkout>]
               [--force-cli] [--upgrade-skills] [--dry-run]

Agents:
  codex        ${CODEX_HOME:-~/.codex}/skills
  claude-code  ~/.claude/skills
  opencode     ${XDG_CONFIG_HOME:-~/.config}/opencode/skills
  cursor       ~/.cursor/skills
  kimi-code    ${KIMI_CODE_HOME:-~/.kimi-code}/skills

Options:
  --agent <name>       Skill host (default: codex)
  --version <tag>      Install the same release tag for CLI and Skills
  --install-dir <dir>  Binary directory (default: ~/.local/bin)
  --skills-dir <dir>   Override the Agent Skill directory
  --source <checkout>  Use installer scripts and Skills from a local checkout
  --force-cli          Explicitly replace an existing CLI binary
  --upgrade-skills     Explicitly upgrade intact, managed Skills
  --dry-run            Show both installation plans without writing
  -h, --help           Show this help

The workflow never registers a Vault and never performs a silent update.
EOF
}

fail() {
  printf 'obs-cli bootstrap: %s\n' "$*" >&2
  exit 1
}

while (($# > 0)); do
  case "$1" in
    --agent)
      (($# >= 2)) || fail "--agent requires a value"
      agent="$2"
      shift 2
      ;;
    --version)
      (($# >= 2)) || fail "--version requires a value"
      version="$2"
      shift 2
      ;;
    --install-dir)
      (($# >= 2)) || fail "--install-dir requires a value"
      install_dir="$2"
      shift 2
      ;;
    --skills-dir)
      (($# >= 2)) || fail "--skills-dir requires a value"
      skills_dir="$2"
      shift 2
      ;;
    --source)
      (($# >= 2)) || fail "--source requires a value"
      source_root="$2"
      shift 2
      ;;
    --force-cli)
      force_cli=true
      shift
      ;;
    --upgrade-skills)
      upgrade_skills=true
      shift
      ;;
    --dry-run)
      dry_run=true
      shift
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      fail "unknown argument: $1"
      ;;
  esac
done

case "$agent" in
  codex | opencode | cursor | custom) ;;
  claude | claude-code) agent="claude-code" ;;
  kimi | kimicode | kimi-code) agent="kimi-code" ;;
  *) fail "unsupported Agent: $agent" ;;
esac
if [[ "$agent" == "custom" && -z "$skills_dir" ]]; then
  fail "--agent custom requires --skills-dir"
fi
[[ "$repository" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] ||
  fail "invalid GitHub repository: $repository"
if [[ -n "$version" &&
  ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$ ]]; then
  fail "invalid release tag: $version"
fi

work_dir=$(mktemp -d "${TMPDIR:-/tmp}/obs-cli-bootstrap.XXXXXX")
trap 'rm -rf "$work_dir"' EXIT

if [[ -n "$source_root" ]]; then
  source_root=$(CDPATH= cd -- "$source_root" && pwd)
  cli_installer="$source_root/scripts/install.sh"
  skill_installer="$source_root/scripts/install-skills.sh"
  [[ -f "$cli_installer" && -f "$skill_installer" ]] ||
    fail "--source is not an obs-cli checkout"
else
  command -v curl >/dev/null 2>&1 ||
    fail "required command not found: curl"
  raw_ref="${version:-main}"
  raw_base="https://raw.githubusercontent.com/${repository}/${raw_ref}/scripts"
  cli_installer="$work_dir/install.sh"
  skill_installer="$work_dir/install-skills.sh"
  curl --fail --location --silent --show-error \
    --output "$cli_installer" "$raw_base/install.sh"
  curl --fail --location --silent --show-error \
    --output "$skill_installer" "$raw_base/install-skills.sh"
fi

cli_args=(--install-dir "$install_dir")
if [[ -n "$version" ]]; then
  cli_args+=(--version "$version")
fi
if [[ "$force_cli" == true ]]; then
  cli_args+=(--force)
fi
if [[ "$dry_run" == true ]]; then
  cli_args+=(--dry-run)
fi

printf '==> Installing obs-cli\n'
bash "$cli_installer" "${cli_args[@]}"

selected_version="$version"
if [[ -z "$selected_version" ]]; then
  if [[ "$dry_run" == true ]]; then
    command -v obs-cli >/dev/null 2>&1 ||
      fail "--dry-run requires --version when obs-cli is not already on PATH"
    selected_version=$(obs-cli --version | awk '{print $3}')
  else
    binary="$install_dir/obs-cli"
    if [[ -x "$install_dir/obs-cli.exe" ]]; then
      binary="$install_dir/obs-cli.exe"
    fi
    [[ -x "$binary" ]] || fail "installed obs-cli binary is not executable"
    selected_version=$("$binary" --version | awk '{print $3}')
  fi
fi

skill_args=(--agent "$agent" --version "$selected_version")
if [[ -n "$skills_dir" ]]; then
  skill_args+=(--dest "$skills_dir")
fi
if [[ -n "$source_root" ]]; then
  skill_args+=(--source "$source_root")
fi
if [[ "$upgrade_skills" == true ]]; then
  skill_args+=(--upgrade)
fi
if [[ "$dry_run" == true ]]; then
  skill_args+=(--dry-run)
fi

printf '==> Installing official Skills for %s\n' "$agent"
bash "$skill_installer" "${skill_args[@]}"

if [[ "$dry_run" == true ]]; then
  printf 'Bootstrap dry-run complete; no CLI or Skill files were changed.\n'
  exit 0
fi

binary="$install_dir/obs-cli"
if [[ -x "$install_dir/obs-cli.exe" ]]; then
  binary="$install_dir/obs-cli.exe"
fi
doctor_args=(doctor --agent "$agent" --output json)
if [[ -n "$skills_dir" ]]; then
  doctor_args+=(--skills-path "$skills_dir")
fi

printf '==> Running offline installation audit\n'
"$binary" "${doctor_args[@]}"

printf 'Bootstrap completed for %s at version %s.\n' "$agent" "$selected_version"
printf 'Vaults were not registered automatically. Next run:\n'
printf '  obs-cli vault discover --output json\n'
printf '  obs-cli vault list --output json\n'
