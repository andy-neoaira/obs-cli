# V1 命名规范与映射表

- 状态：`已冻结并实施`
- 所有者：`obs-cli` 与 `miniobsidian.nvim`
- 生效条件：V1-T01 通过评审

## 1. 命名层级

### 1.1 产品版本

产品遵循 SemVer，Agent-first 首个候选版本为：

```text
v1.0.0-rc.1
```

产品版本与线协议版本独立。未来 `obs-cli v1.3.0` 可以继续实现 `obs-cli/v1`。

### 1.2 线协议

机器可读 envelope 使用：

```text
obs-cli/v1
```

只有发生不兼容的 envelope、错误模型或 operation 语义改变时，才升级为
`obs-cli/v2`。普通新增可选字段或新增 operation 不升级主协议。

### 1.3 Operation

Operation 使用稳定、无产品版本后缀的领域名称：

```text
note.get
note.patch
vault.add
search.content
link.backlinks
```

Capability 中已有的 operation `version` 从 `1` 开始独立演进。

### 1.4 Go 内部名称

Go 文件、类型、函数、变量不携带 `V1` 或 `V2`：

```text
Config
Store
Registry
ConfigPath
newNoteCommand
renderEnvelope
```

只有两个版本必须同时编译且无法通过包边界隔离时，才允许带版本后缀；此情况必须先新增
ADR，不得临时引入。

### 1.5 配置

配置路径和根类型固定为：

```text
obs-cli/config.json
Config
version: 1
```

正式实现不读取 `config-v2.json`，也不提供自动迁移。

### 1.6 Schema

所有公开 JSON Schema 文件必须显式带首版版本：

```text
response-v1.schema.json
capabilities-v1.schema.json
search-audit-report-v1.schema.json
compare-synthesis-report-v1.schema.json
project-status-report-v1.schema.json
```

文件名、`$id`、测试引用和 Skill 引用必须一致。

### 1.7 Capability

Feature flag 表达能力，不表达当前产品历史：

```text
daily_notes
link_inspection
metadata_operations
note_operations
content_search
```

当 operation discovery 已能精确表达能力时，应优先删除重复 flag，而不是机械改名。

## 2. 强制映射

| 当前名称 | 目标名称 |
|---|---|
| `obs-cli/v2` | `obs-cli/v1` |
| `v2.0.0-rc.1` | `v1.0.0-rc.1` |
| `V2Config` | `Config` |
| `NewV2Config` | `NewConfig` |
| `ValidateV2Config` | `ValidateConfig` |
| `V2Path` | `ConfigPath` |
| `ObsCLIV2ConfigFile` | `ConfigFile` |
| `config-v2.json` | `config.json` |
| `v2_store.go` | `store.go` |
| `v2_registry.go` | `registry.go` |
| `renderV2` | `renderEnvelope` |
| `newV2Namespace` | `newNamespace` |
| `newNoteV2Command` | `newNoteCommand` |
| `newVaultV2Command` | `newVaultCommand` |
| `newDailyV2Command` | `newDailyCommand` |
| `newMetadataV2Command` | `newMetadataCommand` |
| `newSearchV2Command` | `newSearchCommand` |
| `newLinkV2Command` | `newLinkCommand` |
| `*_v2.go` | 对应的 `*.go` |
| `*_v2_test.go` | 对应的 `*_test.go` |
| `*-v2.schema.json` | `*-v1.schema.json` |
| `*-v2.json` golden fixture | `*-v1.json` |

## 3. Capability 决策表

| 当前 flag | 目标 | 理由 |
|---|---|---|
| `daily_notes_v2` | 删除 | `daily.get/create/append` 是能力事实源 |
| `link_inspection_v2` | 删除 | `link.backlinks` 是能力事实源 |
| `metadata_v2` | 删除 | `metadata.get/set` 是能力事实源 |
| `note_operations_v2` | 删除 | `note.*` operation 是能力事实源 |
| `search_v2` | 删除 | `search.content` 是能力事实源 |

上述五项直接删除，不改成新的领域 flag。跨 operation 的行为保证继续由现有无版本 flag
表达，并通过实现与规范双向集合测试。

## 4. 历史边界决策

- 唯一路径函数使用 `ConfigPath`。
- 删除 `vault migrate` 及其旧 `preferences.json` 导入实现；新用户使用
  `vault discover`、`vault add` 建立注册表。
- 删除 `docs/agent-first-v2/` 和 V1→V2 迁移叙事，不建立历史归档副本；Git 历史作为
  唯一历史记录。
- 保留许可证、派生关系和第三方通知。

## 5. 必须保留的版本名称

以下名称是独立首版契约，不属于历史包袱：

```text
obs-write/v1
vault-contract/v1
agent-result-v1.schema.json
agent-handoff-v1.schema.json
compatibility-v1.schema.json
miniobsidian.agent-result/v1
miniobsidian.agent-handoff/v1
three-client-e2e/v1
```

第三方 module major version必须原样保留，例如：

```text
github.com/gdamore/tcell/v2
gopkg.in/yaml.v2
```

## 6. 禁止名称

除门禁 allowlist 外，当前产品文件中禁止：

```text
obs-cli/v2
v2.0.0-rc.1
config-v2.json
Agent-first V2
V2Config
new*V2Command
renderV2
*_operations_v2
*_notes_v2
search_v2
metadata_v2
link_inspection_v2
```

旧执行计划和迁移叙事按本规范删除，不建立仓库内归档。需要追溯时使用 Git 历史，不为
历史内容扩大命名门禁 allowlist。

## 7. 评审检查表

- [x] 每个目标名称只有一个含义。
- [x] 内部名称不重复携带协议版本。
- [x] 外部契约均从 V1 开始。
- [x] Capability 不重复 operation discovery。
- [x] 配置不存在双文件或双 Schema。
- [x] `miniobsidian.nvim` 使用同一协议常量。
- [x] 第三方 `/v2` 不被误改。
- [x] 禁止名称可以被自动门禁检测。
