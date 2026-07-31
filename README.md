# obs-cli

Agent-first, non-interactive operations for Obsidian Vaults.

`obs-cli` V1 is a machine-readable execution layer for AI Agents, scenario-oriented Skills, scripts, and editor integrations. It operates on local Markdown files while enforcing Vault boundaries, revision preconditions, atomic writes, dry-run plans, and stable JSON errors.

This project is a derivative of [Yakitrak/notesmd-cli](https://github.com/Yakitrak/notesmd-cli). See [LICENSE](./LICENSE) and [THIRD_PARTY_NOTICES.md](./THIRD_PARTY_NOTICES.md).

## Install

`obs-cli` and its Agent Skills are installed separately. The executable must be
on the `PATH` inherited by the Agent process; a Skill is an instruction package
and does not install or embed the CLI.

### One-command bootstrap

Install the CLI and the same release of all official Skills, then run the
offline installation audit:

```bash
curl -fsSL \
  https://raw.githubusercontent.com/andy-neoaira/obs-cli/main/scripts/bootstrap.sh |
  bash -s -- --agent codex --version v1.0.0-rc.1
```

The first release supports these user-level Skill hosts:

| `--agent` | Default Skill directory |
|---|---|
| `codex` | `${CODEX_HOME:-$HOME/.codex}/skills` |
| `claude-code` | `~/.claude/skills` |
| `opencode` | `${XDG_CONFIG_HOME:-$HOME/.config}/opencode/skills` |
| `cursor` | `~/.cursor/skills` |
| `kimi-code` | `${KIMI_CODE_HOME:-$HOME/.kimi-code}/skills` |

Change only the `--agent` value to target another supported host. The bootstrap
does not register, scan, or modify a Vault. Existing binaries and Skills are
also not silently overwritten: use `--force-cli` and `--upgrade-skills`
explicitly for a managed upgrade. Preview both plans with `--dry-run`.

### 1. Install the CLI

The recommended user installation downloads a checksummed binary from GitHub
Releases. The installer supports macOS, Linux, and Windows shells and installs
to `~/.local/bin` by default:

```bash
curl -fsSLo /tmp/obs-cli-install.sh \
  https://raw.githubusercontent.com/andy-neoaira/obs-cli/main/scripts/install.sh

# Inspect the script before running it.
less /tmp/obs-cli-install.sh
bash /tmp/obs-cli-install.sh
```

Pin a release or choose another user-writable binary directory when required:

```bash
bash /tmp/obs-cli-install.sh \
  --version v1.0.0-rc.1 \
  --install-dir "$HOME/.local/bin"
```

The installer verifies `checksums.txt`, validates the downloaded binary through
`--version` and `capabilities`, and refuses to replace an existing binary unless
`--force` is explicit. Preview platform and destination resolution with
`--dry-run`.

This repository is currently a release candidate. The download command becomes
usable after the requested tag and its GitHub Release assets have been
published. Developers can instead build from the checked-out source:

```bash
go build -mod=vendor -o obs-cli .
mkdir -p "$HOME/.local/bin"
install -m 0755 obs-cli "$HOME/.local/bin/obs-cli"
```

Verify the installation from the same environment that launches the Agent:

```bash
obs-cli capabilities --output json
```

If that command works in an interactive terminal but not from the Agent, add the
installation directory to the Agent process environment rather than assuming it
inherits the terminal's `PATH`.

### 2. Install the Agent Skills separately

The official distributable list is
[`skills/install-manifest.txt`](./skills/install-manifest.txt). It contains the
11 `obsidian-*` Skills; `_template` and `evals` are development resources and
are never installed.

Download the installer and let it select the same tag as the installed CLI:

```bash
curl -fsSLo /tmp/obs-cli-install-skills.sh \
  https://raw.githubusercontent.com/andy-neoaira/obs-cli/main/scripts/install-skills.sh

# Inspect the script before running it.
less /tmp/obs-cli-install-skills.sh
bash /tmp/obs-cli-install-skills.sh --agent codex
```

Replace `codex` with `claude-code`, `opencode`, `cursor`, or `kimi-code` for the
other supported hosts. Start a new Agent session if newly installed Skills are
not detected immediately. Normal installation refuses to overwrite any existing
Skill directory; explicit upgrade additionally verifies managed metadata and
the installed content digest.

When working from a local checkout, install exactly that checkout:

```bash
./scripts/install-skills.sh --agent codex --source .
```

For another Agent that accepts `SKILL.md` packages, provide its skill directory
explicitly:

```bash
./scripts/install-skills.sh \
  --agent custom \
  --dest /absolute/path/to/agent/skills \
  --source .
```

There is no universal skill directory across Agent products. The custom mode
only copies the official packages; whether and when they are discovered is
controlled by the target Agent.

Keep the CLI binary and Skills on the same release tag. Skills negotiate the
actual runtime surface with `obs-cli capabilities --output json --require ...`,
but installing from a moving `main` branch together with an older binary can
still produce an avoidable capability mismatch.

### 3. Audit and upgrade

Run the local audit without network access:

```bash
obs-cli doctor --agent codex --output json
```

It checks the executable, configuration, registered Vault paths, the official
Codex Skills, managed metadata, local Skill modifications, and CLI/Skill version
alignment. Online release lookup is always explicit:

```bash
obs-cli doctor --agent codex --online --output json
obs-cli update check --output json
```

Preview and manually apply a verified CLI upgrade:

```bash
obs-cli update apply --version v1.0.0 --dry-run --output json
obs-cli update apply --version v1.0.0 --output json
```

The apply command downloads the platform asset and `checksums.txt`, validates
the candidate's version and `obs-cli/v1` capabilities, keeps the old executable
as `<path>.previous`, and then replaces it. It never runs automatically. On
Windows, use `scripts/install.sh --version <tag> --force` because a running
executable cannot be replaced safely by the same process.

Skills have a separate explicit upgrade:

```bash
bash /tmp/obs-cli-install-skills.sh \
  --agent codex \
  --version v1.0.0 \
  --upgrade \
  --dry-run

bash /tmp/obs-cli-install-skills.sh \
  --agent codex \
  --version v1.0.0 \
  --upgrade
```

The installer records a digest and managed version at initial installation.
Upgrade stops if metadata is missing or `SKILL.md` has local modifications.
Previous managed Skills are retained in the backup directory printed after a
successful upgrade.

### 4. Register a Vault

Installation does not automatically scan or register personal Vaults. Discover
and inspect candidates first:

```bash
obs-cli vault discover --output json
obs-cli vault list --output json
```

Preview the registry change, then apply the same explicit target:

```bash
obs-cli vault add /absolute/path/to/vault \
  --name Personal \
  --set-default \
  --dry-run \
  --output json

obs-cli vault add /absolute/path/to/vault \
  --name Personal \
  --set-default \
  --output json
```

## Status

The first Agent-first release uses this stable top-level command tree:

```text
capabilities  vault  note  search  metadata
link          daily  doctor  update  template  batch
```

Implemented operations are discovered through `capabilities`. `doctor` performs
an offline audit by default, while `update` only checks or changes versions when
explicitly invoked. Reserved namespaces return `CAPABILITY_UNSUPPORTED`; they
never fall back to interactive picker, GUI, or TTY behavior.

## Build

Requires the Go version declared in `go.mod`.

```bash
go build -mod=vendor -o obs-cli .
./obs-cli capabilities --output json
```

## Agent workflow

Register a Vault:

```bash
./obs-cli vault add /absolute/path/to/vault \
  --name Personal \
  --set-default \
  --dry-run

./obs-cli vault add /absolute/path/to/vault \
  --name Personal \
  --set-default
```

Discover supported operations before using a Skill:

```bash
./obs-cli capabilities \
  --output json \
  --require note.get \
  --require note.patch
```

Create content safely through stdin:

```bash
printf '# Project\n' |
  ./obs-cli note create Projects/demo \
    --vault Personal \
    --content-file -
```

Read the note and retain the returned revision:

```bash
./obs-cli note get Projects/demo \
  --vault Personal \
  --request-id agent-read-001
```

Preview and apply a unique-context patch:

```bash
printf 'Project' > /tmp/obs-cli-match.txt
printf 'Project Alpha' |
  ./obs-cli note patch Projects/demo \
    --vault Personal \
    --match-file /tmp/obs-cli-match.txt \
    --content-file - \
    --if-match 'sha256:<revision-from-get>' \
    --dry-run

printf 'Project Alpha' |
  ./obs-cli note patch Projects/demo \
    --vault Personal \
    --match-file /tmp/obs-cli-match.txt \
    --content-file - \
    --if-match 'sha256:<revision-from-get>'
```

The placeholder revision must be replaced with the exact revision returned by `note get`.

## Implemented V1 operations

```text
capabilities

vault discover
vault list
vault get
vault add
vault remove
vault set-default

note list
note get
note create
note append
note patch
note replace
note delete
note move

daily get
daily create
daily append

metadata get
metadata set

search content
link backlinks
```

All mutating operations support `--dry-run`. Note updates use `sha256:<64 lowercase hex>` revisions. `replace`, `delete`, `patch`, and `move` require `--if-match`; `replace` and `delete` expose an explicit `--unsafe-no-if-match` escape hatch that default Skills must not use.

## Protocol guarantees

- stdout contains one `obs-cli/v1` JSON envelope.
- stderr is diagnostic only and includes the request ID.
- operation names and error codes are stable within V1.
- Vault logical paths reject traversal, hidden paths, and symlink escape.
- create never overwrites.
- append is implemented as revision-aware atomic replacement.
- patch requires exactly one context match.
- move creates the target, rewrites parsed links, and deletes the source in one recoverable transaction.
- dry-run creates no configuration, lock, temporary, recovery, or Vault files.

Specifications:

- [Documentation index](./docs/README.md)
- [Command reference](./docs/COMMAND_REFERENCE.md)
- [CLI protocol](./docs/spec/CLI_PROTOCOL.md)
- [Capabilities and dry-run](./docs/spec/CAPABILITIES.md)
- [Vault path policy](./docs/spec/PATH_POLICY.md)
- [Note operations](./docs/spec/NOTE_OPERATIONS.md)
- [Move transactions](./docs/spec/MOVE_TRANSACTIONS.md)
- [Concurrency and writes](./docs/spec/CONCURRENCY_AND_WRITES.md)

Legacy command aliases, fuzzy pickers, editor launching, Obsidian URI launching, cwd-based Vault selection, and TTY confirmation are not part of the Agent-first command surface.

## Development and release gate

```bash
make release-check
```

The release gate checks formatting, vet, tests, race detection, coverage, protocol schemas, cross-platform builds, licenses, release notices, and the temporary-Vault RC CRUD/conflict smoke test.

Release archives contain the binary, `LICENSE`, `THIRD_PARTY_NOTICES.md`, and vendored dependency license/notice files.
