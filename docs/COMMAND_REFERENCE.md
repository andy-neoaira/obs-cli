# obs-cli V1 命令参考

所有已实现命令默认输出一个 `obs-cli/v1` JSON object。

## 通用参数

| 参数 | 适用范围 |
|---|---|
| `--output json` | 所有 V1 operation |
| `--request-id <id>` | 所有 V1 operation |
| `--vault <id-or-name-or-registered-path>` | Vault-scoped operation |
| `--dry-run` | 修改 operation |
| `--if-match <revision>` | 条件修改 operation |

## Capabilities

```text
capabilities [--require <operation>]...
```

## Vault

```text
vault discover
vault list
vault get <id-or-name>
vault add <path> [--name <name>] [--set-default] [--dry-run]
vault remove <id-or-name> [--dry-run]
vault set-default <id-or-name> [--dry-run]
```

## Note

```text
note list
note get <path>
note create <path> --content-file <file|-> [--dry-run]
note append <path> --content-file <file|-> [--section <heading>] [--if-match <revision>] [--dry-run]
note patch <path> --match-file <file|-> --content-file <file|-> --if-match <revision> [--dry-run]
note replace <path> --content-file <file|-> --if-match <revision> [--dry-run]
note delete <path> --if-match <revision> [--dry-run]
note move <source> <target> --if-match <revision> [--dry-run]
```

`replace/delete` 可以由人类显式传入 `--unsafe-no-if-match`，默认 Agent/Skill 禁止使用。

## Daily、Metadata、Search 与 Link

```text
daily get [--date <YYYY-MM-DD>]
daily create [--date <YYYY-MM-DD>] [--dry-run]
daily append --content-file <file|-> [--date <YYYY-MM-DD>] [--section <heading>] [--if-match <revision>] [--dry-run]

metadata get <path>
metadata set <path> --key <key> --value <value> --if-match <revision> [--dry-run]

search content <query> [--scope <directory>] [--page <n>] [--page-size <n>] [--max-files <n>]
link backlinks <path> [--scope <directory>] [--max-files <n>]
```

## Doctor 与显式升级

```text
doctor [--agent <name>] [--skills-path <absolute-path>] [--online]

update check
update apply [--version <tag>] [--dry-run]
```

`doctor` 默认只执行本地审计，检查 CLI、配置、Vault 注册和所选 Agent 的 Skills。
内置 Agent 名称为 `codex`、`claude-code`、`opencode`、`cursor` 和 `kimi-code`；
其他宿主可以同时传入自定义名称与 `--skills-path`。
只有显式传入 `--online` 才查询 GitHub Releases。

`update check` 只检查版本；`update apply` 才会下载、校验 checksum、验证候选二进制并
替换当前 CLI。它不会自动执行，也不会升级 Skills。Windows 上的二进制替换应继续
使用 `scripts/install.sh --force`。

Skills 使用单独的显式升级流程：

```text
scripts/install-skills.sh --agent codex --version <tag> --upgrade [--dry-run]
```

缺少托管 metadata 或本地 `SKILL.md` 已修改时，升级会停止而不是覆盖。

## 预留命名空间

```text
template  batch
```

在 capabilities 声明具体 operation 前，调用这些命名空间会返回 `CAPABILITY_UNSUPPORTED`（退出码 8）。
