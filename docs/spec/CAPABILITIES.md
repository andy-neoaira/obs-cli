# obs-cli V1 Capability 与 Dry-run 约定

- 协议：`obs-cli/v1`
- Schema：[capabilities-v1.schema.json](./schema/capabilities-v1.schema.json)
- 状态：已实现

## 1. 发现能力

Agent 和 Skill 在首次写入前应执行：

```bash
obs-cli capabilities --output json \
  --require vault.add \
  --require vault.set-default
```

命令成功时返回 CLI 版本、协议版本、写入协议、Vault 共同规范、稳定 operation 列表、feature flags 和平台限制。缺少任意 `--require` operation 时返回 `CAPABILITY_UNSUPPORTED`，退出码为 `8`，调用方必须在写入前停止。

调用方只依赖 operation 名称、版本、声明的通用参数和 feature flag，不解析 `--help` 或自然语言描述。

当前 `vault_contract.target` 与 `vault_contract.implemented` 均为
`vault-contract/v1`。编辑器 Adapter 必须同时检查 `obs-cli/v1`、该共同规范和场景
所需 operation；任一不匹配时只禁用 CLI 高级集成，不影响编辑器自身功能。

## 2. 稳定名称

当前稳定 feature flag：

| 名称 | 含义 |
|---|---|
| `atomic_writes` | 单文件写入通过临时文件和原子替换提交 |
| `json_error_envelopes` | 失败返回稳定 JSON envelope |
| `multi_file_transactions` | 存在带回滚语义的多文件事务内核 |
| `revision_preconditions` | 存储内核支持 revision 条件写入 |
| `vault_discovery_read_only` | Obsidian Vault 发现不会修改官方配置 |
| `vault_path_policy` | Vault 路径执行统一边界和符号链接校验 |
| `dry_run_plans` | 当前 capability 中声明为 mutating 的 operation 支持 dry-run |
| `move_plan_preconditions` | `note.move` dry-run 返回确定性 `plan_hash`，apply 可用 `--if-plan-hash` 绑定已审查的 source、target、链接集合和 revision |

当前 Daily 与 Metadata operation：

| Operation | 修改型 | 关键语义 |
|---|---:|---|
| `daily.get` | 否 | 返回官方配置解析后的路径、存在状态、note 与 revision |
| `daily.create` | 是 | 按官方 folder/format/template 创建，不覆盖现有 Daily |
| `daily.append` | 是 | 携带 revision 向整篇或唯一 section 追加 |
| `note.get` | 否 | 返回正文、Frontmatter、路径、revision、body_revision 和 modified_at |
| `metadata.get` | 否 | 返回目标路径、revision、body_revision 和解析后的 frontmatter |
| `metadata.set` | 是 | 只设置一个键，保留正文和未知字段，要求 revision 并回显 body_revision |
| `search.content` | 否 | 仅扫描 Markdown，限制页大小和读取文件数 |
| `link.backlinks` | 否 | 有界扫描 Wikilink/Markdown link，返回来源证据；允许检查不存在的 target |
| `doctor.audit` | 否 | 默认离线审计 CLI、Vault registry 和指定 Agent 的托管 Skills；`--online` 才查询 Release |
| `update.check` | 否 | 显式查询 GitHub Release，不在普通命令启动时自动执行 |
| `update.apply` | 是 | 显式下载、校验、验证并替换当前二进制；支持 dry-run |

operation 名称和 feature flag 只增不改。新增名称属于兼容变更；删除、改名或改变已有名称的语义需要升级协议主版本。弃用项通过新增的可选 deprecation 元数据公告；调用方必须忽略未知可选字段。

`feature_flags[name] == false` 表示当前版本不提供该完整场景，不能解释为永久不支持。

## 3. 通用参数

CLI 通过同一绑定层注册以下通用参数，具体 operation 是否支持由 `operations[].common_flags` 声明：

- `--output`
- `--request-id`
- `--dry-run`
- `--if-match`
- `--if-plan-hash`（当前仅 `note.move`）
- `--vault`

调用方不得仅因为参数存在于其他命令就假定当前 operation 支持它。

## 4. Dry-run

所有在 capabilities 中声明为 `mutating: true` 的 operation 均支持 `--dry-run`。成功结果固定包含：

```json
{
  "dry_run": true,
  "applied": false,
  "changed": true,
  "plan": {
    "changes": [],
    "risks": [],
    "preconditions": []
  }
}
```

`changes` 使用 `create`、`update`、`move` 或 `delete`，同时标识 `resource` 和 `target`。dry-run 会完成适用的解析、存在性、冲突和安全前置检查，但不会创建配置目录、配置文件、锁、临时文件或审计记录。

Skill 在 dry-run 成功后仍需将 apply 当作独立请求；并发状态可能在两次调用之间改变。
`note.move` 可用 `--if-plan-hash` 把 apply 绑定到已审查的 dry-run 链接集合。

`update.apply --dry-run` 只解析 Release、目标资产、当前可执行文件、备份位置和
checksum 资产，不下载归档、不创建 staging/backup 文件，也不替换二进制。真正的
apply 必须是独立的显式调用；CLI 不提供后台或静默自动更新。

## 5. Skill 场景示例

“注册一个 Vault 并设为默认”的 Skill 应按以下顺序执行：

1. 调用 `capabilities --require vault.add --require vault.set-default`。
2. 调用 `vault add <path> --name <name> --set-default --dry-run`。
3. 检查 envelope、`plan.changes`、`risks` 和 `preconditions`。
4. 获得场景所需确认后，以相同业务参数执行不带 `--dry-run` 的命令。
5. 读取成功 envelope 中返回的稳定 Vault ID，后续调用优先使用该 ID。

如果能力检查失败，Skill 应向 Agent 返回结构化的不支持原因，不得回退到猜测旧命令或直接修改配置文件。
