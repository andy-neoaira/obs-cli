# obs-cli

[🇨🇳 中文](./README_CN.md) | **English**

> **Note**: obs-cli works **without requiring Obsidian to be running**, making it perfect for scripting, automation, and terminal-only environments.

---

## Table of Contents

- [Description](#description)
- [Agent-first V2 Architecture](#agent-first-v2-architecture)
- [Install](#install)
  - [Download Pre-built Binary](#download-pre-built-binary)
  - [Build from Source](#build-from-source)
  - [Headless / No Obsidian Installed](#headless--no-obsidian-installed)
- [Usage](#usage)
  - [Help](#help)
  - [Editor Flag](#editor-flag)
  - [Add Vault](#add-vault)
  - [Remove Vault](#remove-vault)
  - [List Vaults](#list-vaults)
  - [Set Default Vault and Open Type](#set-default-vault-and-open-type)
  - [Open Note](#open-note)
  - [Daily Note](#daily-note)
  - [Search Note](#search-note)
  - [Search Note Content](#search-note-content)
  - [List Vault Contents](#list-vault-contents)
  - [Print Note](#print-note)
  - [Create / Update Note](#create--update-note)
  - [Move / Rename Note](#move--rename-note)
  - [Delete Note](#delete-note)
  - [Frontmatter](#frontmatter)
- [Excluded Files](#excluded-files)
- [Releasing](#releasing)
- [Contribution](#contribution)
- [License](#license)

---

## Description

Obsidian is a powerful and extensible knowledge base application
that works on top of your local folder of plain text notes. This CLI tool (written in Go) will let you interact with the application using the terminal. You are currently able to open, search, list, move, create, update and delete notes.

---

## Agent-first V2 Architecture

The project is evolving toward an Agent-first V2: `obs-cli` will be the safe, machine-readable execution layer used by AI agents and scenario-oriented Skills.

- The Markdown Vault is the single source of truth.
- Obsidian, `obs-cli`, and [`miniobsidian.nvim`](https://github.com/andy-neoaira/miniobsidian.nvim) are peer clients of the same Vault.
- `miniobsidian.nvim` remains fully usable without `obs-cli`; optional future integration is limited to advanced capabilities negotiated through the CLI protocol.
- `obs-cli` does not depend on Neovim, and it only discovers or imports Obsidian-owned configuration without taking ownership of it.
- The V2 core is non-interactive and avoids TTY prompts, built-in pickers, automatic editor launches, and implicit application opening.

See [ADR-001: Agent-first Product Boundary and Three-client Architecture](./docs/architecture/ADR-001-agent-first-boundary.md), the [V2 configuration boundary](./docs/spec/CONFIG_V2.md), and the [V2 execution plan](./docs/agent-first-v2/README.md).

Shared contract status: `target_contract = vault-contract/v1`, `implemented_contract = null`. The target rules are defined in [Vault Conventions](./docs/spec/VAULT_CONVENTIONS.md); conformance will be declared only after the P1 shared fixtures pass.

> The commands documented below describe the current CLI until the V2 migration tasks are completed.

---

## Install

### Download Pre-built Binary

The easiest way to install is to download a pre-built binary from the [GitHub Releases](https://github.com/andy-neoaira/obs-cli/releases) page.

**Supported platforms:**

| OS | Architecture | Release Asset Name |
|---|---|---|
| macOS (Universal) | amd64 + arm64 | `obs-cli_darwin_all.tar.gz` |
| Linux | amd64 | `obs-cli_linux_amd64.tar.gz` |
| Linux | arm64 | `obs-cli_linux_arm64.tar.gz` |
| Windows | amd64 | `obs-cli_windows_amd64.tar.gz` |
| Windows | arm64 | `obs-cli_windows_arm64.tar.gz` |

**One-line install (macOS / Linux):**

```bash
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  darwin) asset_os="darwin"; asset_arch="all" ;;
  linux) asset_os="linux"; asset_arch="$arch" ;;
  *) echo "Unsupported OS: $os" >&2; exit 1 ;;
esac

case "$asset_arch" in
  x86_64) asset_arch="amd64" ;;
  aarch64) asset_arch="arm64" ;;
esac

curl -sL -o obs-cli.tar.gz "https://github.com/andy-neoaira/obs-cli/releases/latest/download/obs-cli_${asset_os}_${asset_arch}.tar.gz"
tar xzf obs-cli.tar.gz

# Move to a directory in your PATH
mkdir -p ~/bin
mv obs-cli ~/bin/
```

> **Note:** On Apple Silicon Macs, `uname -m` prints `arm64`. On Intel Macs, it prints `x86_64`.

**Windows (PowerShell):**

```powershell
# Download latest release
Invoke-WebRequest -Uri "https://github.com/andy-neoaira/obs-cli/releases/latest/download/obs-cli_windows_amd64.tar.gz" -OutFile "obs-cli.tar.gz"

# Extract
tar xzf obs-cli.tar.gz

# Add to PATH (optional)
$env:PATH += ";$PWD"
```

### Build from Source

Requires [Go](https://go.dev/dl/) 1.19 or later.

#### Quick Build

```bash
git clone https://github.com/andy-neoaira/obs-cli.git
cd obs-cli
go build -o obs-cli .
sudo install -m 755 obs-cli /usr/local/bin/
```

#### Using Make

This project includes a `Makefile` with convenient targets for development and release:

```bash
# Build binaries for all supported platforms (Darwin, Linux, Windows)
make build-all

# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Install git hooks
make install-hooks
```

#### Cross-Compilation

Go supports cross-compilation out of the box. Set `GOOS` and `GOARCH` to build for different platforms:

```bash
# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o obs-cli-darwin-amd64 .

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o obs-cli-darwin-arm64 .

# Linux
GOOS=linux GOARCH=amd64 go build -o obs-cli-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -o obs-cli-linux-arm64 .

# Windows
GOOS=windows GOARCH=amd64 go build -o obs-cli-windows-amd64.exe .
```

#### Development Setup

```bash
# 1. Clone the repository
git clone https://github.com/andy-neoaira/obs-cli.git
cd obs-cli

# 2. Download dependencies (using vendor mode)
go mod vendor

# 3. Run tests to ensure everything is working
go test ./...

# 4. Build the binary
go build -o obs-cli .

# 5. Run the CLI
./obs-cli --help
```

#### Install from Source

After building, you can install the binary to your system:

```bash
# Linux / macOS
sudo cp obs-cli /usr/local/bin/

# Or use go install (installs to $GOPATH/bin or $HOME/go/bin)
go install github.com/andy-neoaira/obs-cli@latest

# Verify installation
obs-cli --version
```

### Headless / No Obsidian Installed

If you're running on a headless server or don't have Obsidian installed (e.g., server environments, containers, or systems without a GUI), you can still use this CLI. When Obsidian is installed, it registers vaults automatically. For headless environments, you register them via the CLI instead.

**Setup Instructions:**

```bash
# Register your vault directory
obs-cli add-vault /home/user/vaults/my-brain

# Set it as default
obs-cli set-default-vault "my-brain"

# Or do both in one step
obs-cli add-vault /home/user/vaults/my-brain --set-default
```

For multiple vaults:
```bash
obs-cli add-vault /home/user/vaults/personal
obs-cli add-vault /home/user/vaults/work
obs-cli set-default-vault "personal"
```

You can then pass `--vault "work"` to target a specific vault.

<details>
<summary>Manual setup (without CLI commands)</summary>

1. Create the Obsidian config directory:
   ```bash
   mkdir -p ~/.config/obsidian
   ```

2. Create `obsidian.json` with your vault configuration:
   ```json
   {
     "vaults": {
       "any-unique-id": {
         "path": "/home/user/vaults/my-brain"
       }
     }
   }
   ```
   The key (`any-unique-id`) can be anything. The CLI uses the **directory name** as the vault name (e.g., `my-brain` above). Use the **absolute path**. Do not use `~` as the CLI does not expand it to your home directory.

</details>

---

## Usage

### Help

```bash
# See All command instructions
obs-cli --help
```

### Editor Flag

The `open`, `daily`, `search`, `search-content`, `create`, and `move` commands support the `--editor` (or `-e`) flag, which opens notes in your default text editor instead of the Obsidian application. This is useful for quick edits or when working in a terminal-only environment.

The editor is determined by the `EDITOR` environment variable (e.g., `"vim"`, `"code"`, or `"code -w"`). If not set, it defaults to `vim`.

**Supported editors:**

- Terminal editors: vim, nano, emacs, etc.
- GUI editors with wait flag: VSCode (`code`), Sublime Text (`subl`), Atom, TextMate
  - The CLI automatically adds the `--wait` flag for supported GUI editors to ensure they block until you close the file

**Example:**

```bash
# Set your preferred editor (add to ~/.zshrc or ~/.bashrc to make permanent)
export EDITOR="code"  # or "vim", "nano", "subl", etc.

# Use with supported commands
obs-cli open "note.md" --editor
obs-cli daily --editor
obs-cli search --editor
obs-cli search-content "term" --editor
obs-cli create "note.md" --open --editor
obs-cli move "old.md" "new.md" --open --editor
```

To avoid passing `--editor` every time, configure it as the default open type once:

```bash
obs-cli set-default-vault --open-type editor
```

### Add Vault

Registers a directory as an Obsidian vault. Creates the Obsidian config file (`~/.config/obsidian/obsidian.json`) if it does not exist.

If you have Obsidian installed, vaults are registered automatically when you open them. You only need this command for headless setups or environments where Obsidian is not installed (servers, containers, CI).

```bash
# Register a vault
obs-cli add-vault /path/to/vault

# Register and set as default
obs-cli add-vault /path/to/vault --set-default
```

### Remove Vault

Removes a vault from the Obsidian config by vault name. Does not delete any files on disk. If the removed vault was the default, the default is cleared.

```bash
# Remove by vault name
obs-cli remove-vault "{vault-name}"

```

### List Vaults

Lists all registered Obsidian vaults. The default vault is marked with `(default)`.

```bash
# Lists all vaults (name and path, default marked)
obs-cli list-vaults

# Outputs vaults as JSON
obs-cli list-vaults --json

# Outputs only vault paths (useful for scripting)
obs-cli list-vaults --path-only

# Show only the default vault (name, path, open type)
obs-cli list-vaults --default

# Get just the default vault path (useful for scripting)
obs-cli list-vaults --default --path-only
```

You can add this to your shell configuration file (like `~/.zshrc`) to quickly navigate to the default vault:

```bash
obs_cd() {
    local result=$(obs-cli list-vaults --default --path-only)
    [ -n "$result" ] && cd -- "$result"
}
```

Then you can use `obs_cd` to navigate to the default vault directory within your terminal.

### Set Default Vault and Open Type

Defines the default vault and/or open type for future usage. If no default vault is set, pass `--vault` with other commands to specify which vault to use.

```bash
# Set default vault by name
obs-cli set-default-vault "{vault-name}"

# Set default open type: 'obsidian' (default) or 'editor'
obs-cli set-default-vault --open-type editor

# Set both at once
obs-cli set-default-vault "{vault-name}" --open-type editor
```

When `default_open_type` is set to `editor`, commands that support `--open` will open notes in `$EDITOR` automatically, without needing to pass `--editor` each time.

Note: `open` and other commands in `obs-cli` use this vault's base directory as the working directory, not the current working directory of your terminal.

### Open Note

Open a note in Obsidian or your default editor. The note argument is a vault-relative path; `.md` is optional.

```bash
# Opens note in obsidian vault
obs-cli open "{note-name}"

# Opens note in specified obsidian vault
obs-cli open "{note-name}" --vault "{vault-name}"

# Opens note at a specific heading (case-sensitive)
obs-cli open "{note-name}" --section "{heading-text}"

obs-cli open "{note-name}" --vault "{vault-name}" --section "{heading-text}"

# Opens note in your default editor instead of Obsidian
obs-cli open "{note-name}" --editor
```

### Daily Note

Creates or opens today's daily note directly on disk. **Obsidian does not need to be running**. If `.obsidian/daily-notes.json` exists in the vault, the CLI reads `folder`, `format` (Moment.js date format, default `YYYY-MM-DD`), and `template` from it. A template file's content is used when creating a new daily note. If the config is missing or unreadable, defaults are used (vault root, `YYYY-MM-DD`, no template).

```bash
# Creates / opens daily note in obsidian vault
obs-cli daily

# Creates / opens daily note in specified obsidian vault
obs-cli daily --vault "{vault-name}"

# Creates / opens daily note in your default editor
obs-cli daily --editor

# Adds content to daily note (appends if note already exists)
obs-cli daily --content "abcde"

# Adds multi-line content safely from a file
obs-cli daily --content-file ./daily-entry.md

# Adds content safely from stdin
printf '%s\n' "## Work Log" "- shipped feature" | obs-cli daily --content-file -

# Adds content and opens in editor
obs-cli daily --content "abcde" --editor
```

### Search Note

Starts a fuzzy search displaying notes in the terminal from the vault. You can hit enter on a note to open that in Obsidian.

```bash
# Searches in default obsidian vault
obs-cli search

# Searches in specified obsidian vault
obs-cli search --vault "{vault-name}"

# Searches and opens selected note in your default editor
obs-cli search --editor

```

### Search Note Content

Searches for notes containing a term in note content. By default, it opens an interactive picker and lets you open the selected note in Obsidian (or your editor). For automation and scripting, use `--no-interactive` or `--format json` to print results to stdout.

```bash
# Searches for content in default obsidian vault
obs-cli search-content "search term"

# Searches for content in specified obsidian vault
obs-cli search-content "search term" --vault "{vault-name}"

# Searches and opens selected note in your default editor
obs-cli search-content "search term" --editor

# Prints grep-style results to stdout (non-interactive)
obs-cli search-content "search term" --no-interactive

# Prints JSON for scripts (implies non-interactive mode)
obs-cli search-content "search term" --format json

# Paginated results (default page size: 25, max: 100)
obs-cli search-content "search term" --format json --page 1 --page-size 50

```

### List Vault Contents

Lists files and folders in a vault path. If no path is provided, it lists the vault root.

```bash
# Lists vault root
obs-cli list

# Lists contents of a subfolder in default vault
obs-cli list "001 Notes"

# Lists contents of a subfolder in specified vault
obs-cli list "001 Notes" --vault "{vault-name}"

```

### Print Note

Prints the contents of a note in Obsidian. The note argument is a vault-relative path; `.md` is optional.

```bash
# Prints note in default vault
obs-cli print "{note-name}"

# Prints note by path in default vault
obs-cli print "{note-path}"

# Prints note in specified obsidian
obs-cli print "{note-name}" --vault "{vault-name}"

```

### Create / Update Note

Creates a note directly on disk. **Obsidian does not need to be running**. If the note already exists and neither `--overwrite` nor `--append` is passed, the command fails. Intermediate directories are created automatically.

When the note name has no explicit path (no `/`), the CLI reads `.obsidian/app.json` from the vault to check for a configured default folder (`newFileLocation: "folder"` and `newFileFolderPath`). If configured, the note is placed in that folder. If the config is missing or unreadable, the note is created at the vault root.

```bash
# Creates empty note in default vault
obs-cli create "{note-name}"

# Creates empty note in specified vault
obs-cli create "{note-name}" --vault "{vault-name}"

# Creates note with content
obs-cli create "{note-name}" --content "abcde"

# Creates note with multi-line content from a file
obs-cli create "{note-name}" --content-file ./note.md

# Creates note with content from stdin
printf '%s\n' "# Title" "Body text" | obs-cli create "{note-name}" --content-file -

# Overwrites an existing note
obs-cli create "{note-name}" --content "abcde" --overwrite

# Appends to an existing note
obs-cli create "{note-name}" --content "abcde" --append

# Creates note and opens it in Obsidian
obs-cli create "{note-name}" --content "abcde" --open

# Creates note and opens it in your default editor
obs-cli create "{note-name}" --content "abcde" --open --editor

```

### Move / Rename Note

Moves a given note(path from top level of vault) with new name given (top level of vault). If given same path but different name then its treated as a rename. All links inside vault are updated to match new name.

```bash
# Renames a note in default obsidian
obs-cli move "{current-note-path}" "{new-note-path}"

# Renames a note and given obsidian
obs-cli move "{current-note-path}" "{new-note-path}" --vault "{vault-name}"

# Renames a note in default obsidian and opens it
obs-cli move "{current-note-path}" "{new-note-path}" --open

# Renames a note and opens it in your default editor
obs-cli move "{current-note-path}" "{new-note-path}" --open --editor
```

### Delete Note

Deletes a given note (path from top level of vault).

```bash
# Deletes a note in default vault
obs-cli delete "{note-path}"

# Deletes a note in specified vault
obs-cli delete "{note-path}" --vault "{vault-name}"
```

### Frontmatter

View and modify YAML frontmatter in notes. The note argument is a vault-relative path; `.md` is optional.

```bash
# Print frontmatter of a note
obs-cli frontmatter "{note-name}" --print

# Edit a frontmatter field (creates field if it doesn't exist)
obs-cli frontmatter "{note-name}" --edit --key "status" --value "done"

# Delete a frontmatter field
obs-cli frontmatter "{note-name}" --delete --key "draft"

# Use with a specific vault
obs-cli frontmatter "{note-name}" --print --vault "{vault-name}"
```

## Excluded Files

The CLI respects Obsidian's **Excluded Files** setting (`Settings > Files & Links > Excluded Files`).

- `search` - excluded notes won't appear in the fuzzy finder
- `search-content` - excluded folders won't be searched
- `print` / `frontmatter` - exact vault-relative paths can still access excluded files

Other explicit-path commands (`open`, `move`, `delete`, etc.) can still access excluded files when you provide the exact path.

## Releasing

This project uses [GoReleaser](https://goreleaser.com/) to automatically build and publish binaries to [GitHub Releases](https://github.com/andy-neoaira/obs-cli/releases).

**Trigger a release by pushing a version tag:**

```bash
# 1. Commit your changes
git add .
git commit -m "feat: your changes"

# 2. Tag with a semantic version
git tag v0.1.0

# 3. Push the tag (this triggers the release workflow)
git push origin master
git push origin v0.1.0
```

The GitHub Actions workflow will:
1. Run tests
2. Build binaries for all supported platforms (Darwin, Linux, Windows, amd64 + arm64)
3. Create a GitHub Release with downloadable archives

Only pushes of tags matching `v*.*.*` will trigger a release. Regular pushes to branches will not.

## Contribution

Fork the project, add your feature or fix and submit a pull request. You can also open an [issue](https://github.com/andy-neoaira/obs-cli/issues/new/choose) to report a bug or request a feature.

## Translations

- [English](./README.md)
- [简体中文](./README_CN.md)

## License

This project is derived from [Yakitrak/notesmd-cli](https://github.com/Yakitrak/notesmd-cli). The upstream MIT copyright and permission notice are retained in [LICENSE](./LICENSE), together with the copyright for this project's modifications.

Available under the [MIT License](./LICENSE). Vendored dependency licenses and notices are listed in [THIRD_PARTY_NOTICES.md](./THIRD_PARTY_NOTICES.md) and included in release archives.
