# obs-cli V2 JSON Protocol

- 协议标识：`obs-cli/v2`
- 状态：已接受，等待 P1 实现
- 发布日期：2026-07-24
- Vault 规范：`vault-contract/v1`
- 架构依据：[ADR-001](../architecture/ADR-001-agent-first-boundary.md)
- JSON Schema：[response-v2.schema.json](./schema/response-v2.schema.json)
- Capability 约定：[CAPABILITIES.md](./CAPABILITIES.md)
- Note 原子操作：[NOTE_OPERATIONS.md](./NOTE_OPERATIONS.md)

## 1. 设计目标

`obs-cli/v2` 是供 Agent、Skill、脚本和自动化系统使用的非交互协议。调用方不应解析面向人的 `--help`、表格、颜色或自然语言日志来判断结果。

协议保证：

- 每次非流式调用在 stdout 输出一个 JSON object。
- 成功和失败共享稳定 envelope。
- 领域失败使用稳定错误码，不依赖 message 文本。
- 修改操作可以 dry-run，并通过 revision 实现条件写入。
- 请求和结果可以通过 request ID 关联。

## 2. 传输约定

### 2.1 stdout

JSON 模式下 stdout 必须只包含一个符合 Schema 的 UTF-8 JSON object，末尾可以有一个换行。禁止输出进度条、提示语、ANSI 颜色、表格或额外 JSON 行。

批量结果必须放入单个 envelope 的 `data.items`。`obs-cli/v2` 不定义 NDJSON 流式响应；未来流式协议必须使用独立 capability 和内容类型。

### 2.2 stderr

stderr 只用于诊断信息，可以为空。诊断不得成为判断操作成功与否的唯一依据，不得默认输出笔记正文、密钥、完整用户配置或 Vault 外绝对路径。

stderr 日志应包含 request ID。`--quiet` 可以关闭非必要诊断，但不能改变 stdout envelope。

### 2.3 字符编码

参数、stdin、stdout 和 stderr 使用 UTF-8。JSON 中不得输出非法 UTF-8 字节；无法解码的文件内容通过领域错误报告。

## 3. 通用 Envelope

所有响应必须包含：

| 字段 | 类型 | 说明 |
|---|---|---|
| `protocol_version` | string | 固定为 `obs-cli/v2` |
| `ok` | boolean | 操作是否成功 |
| `operation` | string | 稳定操作名，如 `note.get` |
| `request_id` | string | 调用方提供或 CLI 生成的请求 ID |
| `warnings` | array | 非致命、机器可读警告；无警告时为空数组 |

### 3.1 成功响应

成功响应必须包含 `data`，不得包含 `error`：

```json
{
  "protocol_version": "obs-cli/v2",
  "ok": true,
  "operation": "note.get",
  "request_id": "req-01JXYZ",
  "data": {
    "path": "Projects/demo.md",
    "note_id": "Projects/demo",
    "revision": "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "content": "# Demo\n"
  },
  "warnings": []
}
```

成功但没有发生内容变化仍返回 `ok: true`，并在具体操作数据中返回 `changed: false`。可重复删除不存在目标等行为是否成功由具体 capability 定义，不能仅由 renderer 决定。

### 3.2 失败响应

失败响应必须包含 `error`，不得包含 `data`：

```json
{
  "protocol_version": "obs-cli/v2",
  "ok": false,
  "operation": "note.replace",
  "request_id": "req-01JXYZ",
  "error": {
    "code": "REVISION_CONFLICT",
    "message": "The note changed after it was read.",
    "retryable": true,
    "details": {
      "path": "Projects/demo.md",
      "expected_revision": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      "actual_revision": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
    }
  },
  "warnings": []
}
```

`message` 面向诊断，可以改善措辞或本地化；调用方只能依赖 `code` 和结构化 `details`。

### 3.3 Warning

warning 结构：

| 字段 | 类型 | 说明 |
|---|---|---|
| `code` | string | 稳定或 capability 定义的警告码 |
| `message` | string | 人类可读说明 |
| `details` | object | 可选结构化上下文 |

warning 不改变 `ok`。如果风险使结果不再可信，必须返回失败而不是 warning。

## 4. Operation 命名

operation 使用小写点分命名：

```text
capabilities.get
vault.discover
vault.list
note.get
note.create
note.patch
link.rewrite
daily.append
```

命令别名、人类显示名称和 Cobra 层级不得改变 operation。破坏 operation 语义必须升级协议主版本。

## 5. Request ID

- 通用参数为 `--request-id <id>`。
- 调用方未提供时，CLI 必须生成不可预测且进程内唯一的 ID。
- ID 长度为 1–128，只允许 ASCII 字母、数字、`.`、`_`、`:`、`-`。
- CLI 必须原样返回合法的调用方 ID。
- request ID 用于关联与幂等记录，但不自动等价于幂等键。
- 重试同一业务写入时，Skill 应复用原 request ID，并遵循具体操作的幂等规则。

## 6. 通用参数

### 6.1 `--output`

```text
--output json
```

Agent 核心命令必须支持 JSON。V2 是否将 JSON 设为默认由 P1 命令树决定，但 Skills 必须显式传入 `--output json`，直到 capabilities 声明默认值。

### 6.2 `--vault`

`--vault` 接受稳定 Vault ID 或已注册名称。机器协议不得根据 cwd 隐式选择写入 Vault。未提供时只有存在明确默认 Vault 才可继续，并必须在响应中返回最终 Vault 身份。

### 6.3 `--dry-run`

所有修改操作必须支持 `--dry-run`：

- 执行与真实操作相同的参数、权限、路径和 revision 前置检查。
- 返回 `data.dry_run: true`、`data.applied: false` 和结构化 `plan`。
- 不得创建目录、临时文件、锁文件、配置或审计记录。
- dry-run 成功不保证未来 apply 成功；apply 必须重新检查 revision。
- 计划必须列出所有预期文件变化和无法确定的风险。

### 6.4 `--if-match`

更新、替换、移动和删除通过 `--if-match <revision>` 提供前置条件。revision 语法由并发规范冻结，V2 目标格式为：

```text
sha256:<64 lowercase hex characters>
```

不匹配返回 `REVISION_CONFLICT`。调用方不得在未重新读取和重新规划时去掉前置条件重试。

多文件计划在 `plan.changes[*].expected_revision` 中携带每个文件的 revision。

### 6.5 stdin 与内容文件

多行 Markdown 和不受信任文本必须通过 stdin 或显式内容文件传入，不应拼接进 shell 命令。`-` 表示 stdin 时，一个命令只能有一个 stdin 消费者。

## 7. Dry-run Plan

修改操作的 `data` 至少包含：

```json
{
  "dry_run": true,
  "applied": false,
  "changed": true,
  "plan": {
    "changes": [
      {
        "action": "update",
        "resource": "note",
        "target": "Projects/demo.md",
        "details": {
          "expected_revision": "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
        }
      }
    ],
    "risks": [],
    "preconditions": []
  }
}
```

允许的基础 action 为 `create`、`update`、`move`、`delete`。具体命令可以增加字段，但不得改变基础字段含义。

Note 等 Vault 内资源的 `target` 必须是 Vault 逻辑路径。Vault 注册表操作可以在 `details.canonical_path` 返回待注册的 Vault 根路径；默认不得暴露其他 Vault 外绝对路径。

## 8. 稳定错误码

### 8.1 首批错误码

| code | 含义 | 默认 retryable | 退出码 |
|---|---|---:|---:|
| `INVALID_ARGUMENT` | 参数缺失、冲突或格式错误 | false | 2 |
| `INVALID_FRONTMATTER` | Frontmatter 无法安全解析或修改 | false | 2 |
| `VAULT_NOT_FOUND` | Vault 不存在或不可解析 | false | 3 |
| `NOTE_NOT_FOUND` | Note 不存在 | false | 3 |
| `ALREADY_EXISTS` | create/move 目标已存在 | false | 4 |
| `AMBIGUOUS_NOTE` | 输入对应多个候选 | false | 4 |
| `REVISION_CONFLICT` | 当前 revision 与前置条件不一致 | true | 4 |
| `PATH_OUTSIDE_VAULT` | 路径或符号链接逃逸 Vault | false | 5 |
| `PARTIAL_FAILURE` | 多文件操作未能完整应用或回滚 | false | 7 |
| `CAPABILITY_UNSUPPORTED` | 当前版本或平台不支持所需能力 | false | 8 |
| `INTERNAL_ERROR` | 未分类内部错误或 I/O 失败 | false | 10 |

`retryable: true` 表示调用方在改变前置状态后可以重试，不表示可以原参数立即无限重试。`REVISION_CONFLICT` 必须先重新读取和重新规划。

### 8.2 错误 details 最低要求

- `AMBIGUOUS_NOTE`：返回 `query` 和 `candidates[]`。
- `REVISION_CONFLICT`：返回 `path`、`expected_revision`、`actual_revision`。
- `PATH_OUTSIDE_VAULT`：返回用户提供的逻辑输入和违反的 Vault 规则 ID，不返回 Vault 外真实目标。
- `PARTIAL_FAILURE`：返回 `completed[]`、`failed[]`、`rolled_back[]`、`recovery_actions[]`。
- `CAPABILITY_UNSUPPORTED`：返回 `required` 和当前 capability/version。

### 8.3 新错误码

次版本可以增加错误码，但不能改变既有 code 的含义。调用方遇到未知 code 时应将其视为失败，保留 `details`，不得自动强制重试。

## 9. 退出码

| 退出码 | 类别 |
|---:|---|
| 0 | 成功，包括 dry-run 和无变化 |
| 2 | 调用或内容格式错误 |
| 3 | 目标不存在 |
| 4 | 状态冲突、歧义或目标已存在 |
| 5 | 安全策略拒绝 |
| 6 | 文件系统、配置或外部依赖 I/O 错误 |
| 7 | 部分失败，需要恢复 |
| 8 | capability/平台不支持 |
| 10 | 未分类内部错误 |

退出码用于粗粒度 shell 控制，精确分支必须读取 JSON `error.code`。内部 panic 应被根命令恢复为退出码 10 的协议错误，同时将堆栈限制在显式 debug 诊断中。

## 10. Capability 协商

Agent 在执行场景前应调用：

```bash
obs capabilities --output json
```

`capabilities.get` 至少返回：

- CLI 版本。
- `protocol_versions`。
- `vault_contract` 的 target/implemented。
- operation 列表及版本。
- 通用参数支持情况。
- 平台限制。

如果 Skill 所需 operation、协议或规范不满足，必须在写入前停止并返回 `CAPABILITY_UNSUPPORTED`。

## 11. 兼容性

- 删除必填字段、改变字段类型或改变错误码含义，需要升级 `obs-cli/v3`。
- 新增可选字段可以保持 V2。
- 调用方必须忽略未知可选字段。
- 响应 Schema 的 `additionalProperties` 允许 operation data 扩展，但 envelope 保留字段不得重定义。
- V1 人类输出和 V2 JSON 不共享兼容承诺。

## 12. 安全要求

- JSON 模式不得通过 shell 求值生成参数。
- Vault 路径必须符合 `vault-contract/v1`。
- 失败响应不得泄露 Vault 外路径或敏感环境变量。
- 诊断日志不得默认记录完整笔记正文。
- dry-run 必须真正无写入副作用。
- 任何绕过 revision 的危险操作必须是独立 capability，且不得被默认 Skill 使用。
