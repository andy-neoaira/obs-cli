# obs-cli V2 命令参考

所有已实现命令默认输出一个 `obs-cli/v2` JSON object。

## 通用参数

| 参数 | 适用范围 |
|---|---|
| `--output json` | 所有 V2 operation |
| `--request-id <id>` | 所有 V2 operation |
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
vault migrate [--dry-run]
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

## 预留命名空间

```text
search  metadata  link  daily  template  batch  doctor
```

在 capabilities 声明具体 operation 前，调用这些命名空间会返回 `CAPABILITY_UNSUPPORTED`（退出码 8）。
