#!/usr/bin/env bash

set -euo pipefail

repository="${OBS_CLI_REPOSITORY:-andy-neoaira/obs-cli}"
version="latest"
install_dir="${OBS_CLI_INSTALL_DIR:-${HOME}/.local/bin}"
force=false
dry_run=false

usage() {
  cat <<'EOF'
Install obs-cli from a GitHub Release.

Usage:
  install.sh [--version <tag>] [--install-dir <dir>] [--force] [--dry-run]

Options:
  --version <tag>       Release tag to install (default: latest)
  --install-dir <dir>   Binary directory (default: ~/.local/bin)
  --force               Replace an existing obs-cli binary
  --dry-run             Print the resolved download and destination only
  -h, --help            Show this help

Environment:
  OBS_CLI_REPOSITORY    GitHub owner/repository (default: andy-neoaira/obs-cli)
  OBS_CLI_INSTALL_DIR   Default installation directory
EOF
}

fail() {
  printf 'obs-cli installer: %s\n' "$*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 ||
    fail "required command not found: $1"
}

while (($# > 0)); do
  case "$1" in
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
    --force)
      force=true
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

if [[ "$version" != "latest" &&
  ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$ ]]; then
  fail "invalid release tag: $version"
fi

case "$(uname -s)" in
  Darwin)
    platform="darwin"
    release_arch="all"
    archive_ext="tar.gz"
    binary_name="obs-cli"
    ;;
  Linux)
    platform="linux"
    case "$(uname -m)" in
      x86_64 | amd64) release_arch="amd64" ;;
      arm64 | aarch64) release_arch="arm64" ;;
      *) fail "unsupported Linux architecture: $(uname -m)" ;;
    esac
    archive_ext="tar.gz"
    binary_name="obs-cli"
    ;;
  MINGW* | MSYS* | CYGWIN*)
    platform="windows"
    case "$(uname -m)" in
      x86_64 | amd64) release_arch="amd64" ;;
      arm64 | aarch64) release_arch="arm64" ;;
      *) fail "unsupported Windows architecture: $(uname -m)" ;;
    esac
    archive_ext="zip"
    binary_name="obs-cli.exe"
    ;;
  *)
    fail "unsupported operating system: $(uname -s)"
    ;;
esac

asset="obs-cli_${platform}_${release_arch}.${archive_ext}"
if [[ "$version" == "latest" ]]; then
  release_url="https://github.com/${repository}/releases/latest/download"
else
  release_url="https://github.com/${repository}/releases/download/${version}"
fi
archive_url="${release_url}/${asset}"
checksum_url="${release_url}/checksums.txt"
destination="${install_dir}/${binary_name}"

printf 'obs-cli installer plan:\n'
printf '  release:     %s\n' "$version"
printf '  asset:       %s\n' "$archive_url"
printf '  destination: %s\n' "$destination"

if [[ "$dry_run" == true ]]; then
  exit 0
fi

if [[ -e "$destination" && "$force" != true ]]; then
  fail "$destination already exists; pass --force to replace it"
fi

require_command curl
if [[ "$archive_ext" == "zip" ]]; then
  require_command unzip
else
  require_command tar
fi

work_dir=$(mktemp -d "${TMPDIR:-/tmp}/obs-cli-install.XXXXXX")
trap 'rm -rf "$work_dir"' EXIT

curl --fail --location --silent --show-error \
  --output "$work_dir/$asset" "$archive_url"
curl --fail --location --silent --show-error \
  --output "$work_dir/checksums.txt" "$checksum_url"

expected_checksum=$(
  awk -v asset="$asset" '
    $2 == asset || $2 == "*" asset { print $1 }
  ' "$work_dir/checksums.txt"
)
[[ -n "$expected_checksum" ]] ||
  fail "checksums.txt does not contain $asset"

if command -v sha256sum >/dev/null 2>&1; then
  actual_checksum=$(sha256sum "$work_dir/$asset" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  actual_checksum=$(shasum -a 256 "$work_dir/$asset" | awk '{print $1}')
elif command -v openssl >/dev/null 2>&1; then
  actual_checksum=$(openssl dgst -sha256 "$work_dir/$asset" | awk '{print $NF}')
else
  fail "sha256sum, shasum, or openssl is required to verify the download"
fi

[[ "$actual_checksum" == "$expected_checksum" ]] ||
  fail "checksum mismatch for $asset"

mkdir -p "$work_dir/unpacked"
if [[ "$archive_ext" == "zip" ]]; then
  unzip -q "$work_dir/$asset" -d "$work_dir/unpacked"
else
  tar -xzf "$work_dir/$asset" -C "$work_dir/unpacked"
fi

binary_candidates=()
while IFS= read -r candidate_path; do
  binary_candidates+=("$candidate_path")
done < <(find "$work_dir/unpacked" -type f -name "$binary_name" -print)
[[ "${#binary_candidates[@]}" -eq 1 ]] ||
  fail "release archive must contain exactly one $binary_name"
candidate="${binary_candidates[0]}"
chmod 0755 "$candidate"

version_output=$("$candidate" --version)
if [[ "$version" != "latest" &&
  "$version_output" != "obs-cli version $version" ]]; then
  fail "binary version mismatch: $version_output"
fi
"$candidate" capabilities --output json >/dev/null

mkdir -p "$install_dir"
staged_binary=$(mktemp "${install_dir}/.obs-cli.XXXXXX")
cp "$candidate" "$staged_binary"
chmod 0755 "$staged_binary"
mv -f "$staged_binary" "$destination"

printf 'Installed %s to %s\n' "$version_output" "$destination"
case ":${PATH}:" in
  *":${install_dir}:"*) ;;
  *)
    printf 'Add %s to PATH before using obs-cli.\n' "$install_dir"
    ;;
esac
printf 'Verify with: obs-cli capabilities --output json\n'
