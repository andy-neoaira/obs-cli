# Skill Capability / Version 兼容矩阵

所有场景 Skill 的最低兼容版本为 `obs-cli v1.0.0-rc.1`，协议为
`obs-cli/v1`。`dev` 仅代表同一源码树中的本地开发构建；发布包必须提供可比较的
SemVer。

| Skill | 必需 capability | 修改型 capability |
|---|---|---|
| `obsidian-vault-setup` | `vault.discover`, `vault.list`, `vault.add`, `vault.set-default` | `vault.add`, `vault.set-default` |
| `obsidian-capture` | `vault.get`, `note.get`, `note.create`, `note.append` | `note.create`, `note.append` |
| `obsidian-daily-log` | `daily.get`, `daily.create`, `daily.append` | `daily.create`, `daily.append` |
| `obsidian-project-note` | `note.get`, `note.create`, `note.append`, `metadata.get`, `metadata.set` | `note.create`, `note.append`, `metadata.set` |
| `obsidian-knowledge-search` | `vault.get`, `search.content`, `note.get`, `link.backlinks` | 无 |
| `obsidian-vault-audit` | `vault.get`, `note.list`, `note.get`, `search.content`, `link.backlinks` | 无 |
| `obsidian-inbox-triage` | `vault.get`, `note.list`, `note.get`, `note.move`, `metadata.get`, `metadata.set`, `link.backlinks` | `note.move`, `metadata.set` |
| `obsidian-compare-notes` | `vault.get`, `note.get`, `search.content` | 无 |
| `obsidian-knowledge-synthesis` | `vault.get`, `note.get`, `note.create`, `note.patch` | `note.create`, `note.patch` |
| `obsidian-project-status` | `vault.get`, `note.get`, `daily.get`, `search.content`, `note.append` | `note.append` |
| `obsidian-safe-note-update` | `vault.get`, `note.get`, `note.patch` | `note.patch` |

修改型 capability 仅在用户明确授权写入后要求。运行时仍必须以
`obs-cli capabilities --output json --require ...` 的结果为准；缺失任一当前路径所需
capability 时返回 `CAPABILITY_UNSUPPORTED`，不得回退到历史命令或直接文件写入。
