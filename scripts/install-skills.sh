#!/usr/bin/env bash

set -euo pipefail

repository="${OBS_CLI_REPOSITORY:-andy-neoaira/obs-cli}"
agent="codex"
version=""
source_root=""
destination_root=""
dry_run=false
upgrade=false

script_dir=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
checkout_root=$(CDPATH= cd -- "$script_dir/.." && pwd)

usage() {
  cat <<'EOF'
Install the official obs-cli Skills into an Agent skill directory.

Usage:
  install-skills.sh [--agent codex] [--version <tag>]
                    [--source <checkout>] [--dest <dir>] [--upgrade] [--dry-run]

Options:
  --agent <name>       codex, claude-code, opencode, cursor, kimi-code, or custom
  --version <tag>      Git tag to download (default: latest GitHub Release)
  --source <checkout>  Install from a local obs-cli checkout instead of GitHub
  --dest <dir>         Override the Agent skill directory
  --upgrade            Upgrade intact Skills installed by this installer
  --dry-run            Show the selected Skills and destination without writing
  -h, --help           Show this help

Normal installation never overwrites an existing Skill directory. Upgrade mode
stops if managed metadata is missing or SKILL.md was locally modified. _template
and evals are development resources and are never installed.
EOF
}

fail() {
  printf 'obs-cli Skill installer: %s\n' "$*" >&2
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
    --source)
      (($# >= 2)) || fail "--source requires a value"
      source_root="$2"
      shift 2
      ;;
    --dest)
      (($# >= 2)) || fail "--dest requires a value"
      destination_root="$2"
      shift 2
      ;;
    --dry-run)
      dry_run=true
      shift
      ;;
    --upgrade)
      upgrade=true
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
  codex)
    if [[ -z "$destination_root" ]]; then
      destination_root="${CODEX_HOME:-${HOME}/.codex}/skills"
    fi
    ;;
  claude | claude-code)
    agent="claude-code"
    if [[ -z "$destination_root" ]]; then
      destination_root="${HOME}/.claude/skills"
    fi
    ;;
  opencode)
    if [[ -z "$destination_root" ]]; then
      destination_root="${XDG_CONFIG_HOME:-${HOME}/.config}/opencode/skills"
    fi
    ;;
  cursor)
    if [[ -z "$destination_root" ]]; then
      destination_root="${HOME}/.cursor/skills"
    fi
    ;;
  kimi | kimicode | kimi-code)
    agent="kimi-code"
    if [[ -z "$destination_root" ]]; then
      destination_root="${KIMI_CODE_HOME:-${HOME}/.kimi-code}/skills"
    fi
    ;;
  custom)
    [[ -n "$destination_root" ]] ||
      fail "--agent custom requires --dest"
    ;;
  *)
    fail "unsupported Agent: $agent (supported: codex, claude-code, opencode, cursor, kimi-code, custom)"
    ;;
esac

[[ "$repository" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] ||
  fail "invalid GitHub repository: $repository"
if [[ -n "$version" && "$version" != "latest" &&
  ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$ ]]; then
  fail "invalid release tag: $version"
fi

work_dir=$(mktemp -d "${TMPDIR:-/tmp}/obs-cli-skills-install.XXXXXX")
trap 'rm -rf "$work_dir"' EXIT

if [[ -n "$source_root" ]]; then
  source_root=$(CDPATH= cd -- "$source_root" && pwd)
elif [[ -f "$checkout_root/skills/install-manifest.txt" ]]; then
  source_root="$checkout_root"
  if [[ -n "$version" ]]; then
    printf 'Using local checkout with managed version %s.\n' "$version"
  fi
else
  if [[ -z "$version" || "$version" == "latest" ]]; then
    command -v curl >/dev/null 2>&1 ||
      fail "required command not found: curl"
    latest_url=$(
      curl --fail --location --silent --show-error \
        --output /dev/null \
        --write-out '%{url_effective}' \
        "https://github.com/${repository}/releases/latest"
    )
    version="${latest_url%/}"
    version="${version##*/}"
    printf 'Resolved latest obs-cli Release: %s\n' "$version"
  fi
  [[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$ ]] ||
    fail "invalid or development CLI version: $version"
  command -v curl >/dev/null 2>&1 ||
    fail "required command not found: curl"
  command -v tar >/dev/null 2>&1 ||
    fail "required command not found: tar"

  source_url="https://github.com/${repository}/archive/refs/tags/${version}.tar.gz"
  printf 'Downloading Skills from %s\n' "$source_url"
  curl --fail --location --silent --show-error \
    --output "$work_dir/source.tar.gz" "$source_url"
  mkdir -p "$work_dir/source"
  tar -xzf "$work_dir/source.tar.gz" -C "$work_dir/source"
  source_candidates=()
  while IFS= read -r source_candidate; do
    source_candidates+=("$source_candidate")
  done < <(
    find "$work_dir/source" -mindepth 1 -maxdepth 1 -type d -print
  )
  [[ "${#source_candidates[@]}" -eq 1 ]] ||
    fail "source archive must contain exactly one repository directory"
  source_root="${source_candidates[0]}"
fi

managed_version="$version"
if [[ -z "$managed_version" ]]; then
  managed_version="dev"
  if command -v git >/dev/null 2>&1; then
    exact_tag=$(git -C "$source_root" describe --tags --exact-match 2>/dev/null || true)
    if [[ "$exact_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$ ]]; then
      managed_version="$exact_tag"
    fi
  fi
fi

manifest="$source_root/skills/install-manifest.txt"
[[ -f "$manifest" ]] ||
  fail "official Skill manifest not found: $manifest"

skill_names=()
while IFS= read -r skill_name; do
  [[ -n "$skill_name" && "${skill_name:0:1}" != "#" ]] || continue
  [[ "$skill_name" =~ ^obsidian-[a-z0-9-]+$ ]] ||
    fail "invalid Skill name in manifest: $skill_name"
  [[ -f "$source_root/skills/$skill_name/SKILL.md" ]] ||
    fail "Skill entry point not found: skills/$skill_name/SKILL.md"
  skill_names+=("$skill_name")
done <"$manifest"
[[ "${#skill_names[@]}" -gt 0 ]] ||
  fail "official Skill manifest is empty"

printf 'obs-cli Skill installer plan:\n'
printf '  agent:       %s\n' "$agent"
printf '  source:      %s\n' "$source_root"
printf '  destination: %s\n' "$destination_root"
printf '  version:     %s\n' "$managed_version"
printf '  mode:        %s\n' "$([[ "$upgrade" == true ]] && printf upgrade || printf install)"
printf '  Skills:\n'
for skill_name in "${skill_names[@]}"; do
  printf '    - %s\n' "$skill_name"
done

existing=()
for skill_name in "${skill_names[@]}"; do
  if [[ -e "$destination_root/$skill_name" ]]; then
    existing+=("$skill_name")
  fi
done
if [[ "$upgrade" != true && "${#existing[@]}" -gt 0 ]]; then
  printf 'Existing Skill directories:\n' >&2
  printf '  - %s\n' "${existing[@]}" >&2
  fail "no files were changed; remove or relocate existing directories first"
fi

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print "sha256:" $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print "sha256:" $1}'
  elif command -v openssl >/dev/null 2>&1; then
    openssl dgst -sha256 "$1" | awk '{print "sha256:" $NF}'
  else
    fail "sha256sum, shasum, or openssl is required"
  fi
}

metadata_field() {
  sed -n "s/^[[:space:]]*\"$2\":[[:space:]]*\"\\([^\"]*\\)\"[,]*$/\\1/p" "$1"
}

if [[ "$upgrade" == true ]]; then
  for skill_name in "${skill_names[@]}"; do
    target="$destination_root/$skill_name"
    [[ -d "$target" ]] ||
      fail "upgrade requires every managed Skill; missing: $skill_name"
    metadata="$target/.obs-cli-managed.json"
    [[ -f "$metadata" ]] ||
      fail "$skill_name has no managed metadata; refusing to overwrite it"
    [[ "$(metadata_field "$metadata" managed_by)" == "obs-cli" ]] ||
      fail "$skill_name is not managed by obs-cli"
    [[ "$(metadata_field "$metadata" skill)" == "$skill_name" ]] ||
      fail "$skill_name managed metadata identity is invalid"
    recorded_digest=$(metadata_field "$metadata" skill_digest)
    current_digest=$(sha256_file "$target/SKILL.md")
    [[ -n "$recorded_digest" && "$recorded_digest" == "$current_digest" ]] ||
      fail "$skill_name has local modifications; refusing to overwrite it"
  done
fi

if [[ "$dry_run" == true ]]; then
  exit 0
fi

mkdir -p "$destination_root"
stage_root=$(mktemp -d "${destination_root}/.obs-cli-skills.XXXXXX")
for skill_name in "${skill_names[@]}"; do
  mkdir -p "$stage_root/$skill_name"
  cp -R "$source_root/skills/$skill_name/." "$stage_root/$skill_name/"
  skill_digest=$(sha256_file "$stage_root/$skill_name/SKILL.md")
  printf '{\n  "managed_by": "obs-cli",\n  "version": "%s",\n  "source": "%s",\n  "skill": "%s",\n  "skill_digest": "%s"\n}\n' \
    "$managed_version" "$repository" "$skill_name" "$skill_digest" \
    >"$stage_root/$skill_name/.obs-cli-managed.json"
done

backup_root=""
if [[ "$upgrade" == true ]]; then
  backup_root=$(mktemp -d "${destination_root}/.obs-cli-skills-backup.XXXXXX")
  moved=()
  for skill_name in "${skill_names[@]}"; do
    if ! mv "$destination_root/$skill_name" "$backup_root/$skill_name"; then
      for moved_name in "${moved[@]}"; do
        mv "$backup_root/$moved_name" "$destination_root/$moved_name"
      done
      fail "could not stage existing Skills for upgrade"
    fi
    moved+=("$skill_name")
  done
fi

activated=()
for skill_name in "${skill_names[@]}"; do
  if [[ -e "$destination_root/$skill_name" ]] ||
    ! mv "$stage_root/$skill_name" "$destination_root/$skill_name"; then
    for activated_name in "${activated[@]}"; do
      rm -rf "$destination_root/$activated_name"
    done
    if [[ -n "$backup_root" ]]; then
      for restore_name in "${skill_names[@]}"; do
        if [[ -d "$backup_root/$restore_name" ]]; then
          mv "$backup_root/$restore_name" "$destination_root/$restore_name"
        fi
      done
    fi
    fail "Skill activation failed; previous managed Skills were restored"
  fi
  activated+=("$skill_name")
done
rm -rf "$stage_root"

if [[ "$upgrade" == true ]]; then
  printf 'Upgraded %d Skills in %s\n' "${#skill_names[@]}" "$destination_root"
  printf 'Previous managed Skills are backed up at %s\n' "$backup_root"
else
  printf 'Installed %d Skills into %s\n' "${#skill_names[@]}" "$destination_root"
fi
printf 'Restart or start a new %s session if the Skills are not detected immediately.\n' "$agent"
printf 'Verify that the Agent can run: obs-cli capabilities --output json\n'
