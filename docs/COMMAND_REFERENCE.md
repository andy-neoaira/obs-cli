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

## 预留命名空间

```text
template  batch  doctor
```

在 capabilities 声明具体 operation 前，调用这些命名空间会返回 `CAPABILITY_UNSUPPORTED`（退出码 8）。
