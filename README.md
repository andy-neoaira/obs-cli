# obs-cli

Agent-first, non-interactive operations for Obsidian Vaults.

`obs-cli` V1 is a machine-readable execution layer for AI Agents, scenario-oriented Skills, scripts, and editor integrations. It operates on local Markdown files while enforcing Vault boundaries, revision preconditions, atomic writes, dry-run plans, and stable JSON errors.

This project is a derivative of [Yakitrak/notesmd-cli](https://github.com/Yakitrak/notesmd-cli). See [LICENSE](./LICENSE) and [THIRD_PARTY_NOTICES.md](./THIRD_PARTY_NOTICES.md).

## Status

The first Agent-first release uses this stable top-level command tree:

```text
capabilities  vault  note  search  metadata
link          daily  template  batch  doctor
```

Implemented operations are discovered through `capabilities`. `vault`, `note`, `daily`, `metadata`, `search`, and `link` expose the current V1 operations. Other reserved namespaces return `CAPABILITY_UNSUPPORTED`; they never fall back to interactive picker, GUI, or TTY behavior.

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
