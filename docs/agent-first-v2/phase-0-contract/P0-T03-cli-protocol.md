# P0-T03：CLI JSON 协议与错误模型

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli`
- 依赖：P0-T01

## 目标

定义供 Agent 和 Skill 使用的版本化 JSON 协议、退出码和稳定错误码。

## 实施步骤

1. 新建 `docs/spec/CLI_PROTOCOL.md`，协议标记为 `obs-cli/v2`。
2. 定义成功响应 envelope：`ok`、`operation`、`request_id`、`data`、`warnings`。
3. 定义失败响应 envelope：`ok`、`error.code`、`message`、`details`、`retryable`。
4. 定义 stdout 只输出数据、stderr 只输出诊断信息。
5. 定义退出码与错误类别的映射。
6. 定义 `--output json`、`--request-id`、`--dry-run`、`--if-match` 的通用行为。
7. 创建 JSON Schema 和黄金样例。
8. 冻结第一批错误码。

## 首批错误码

`VAULT_NOT_FOUND`、`NOTE_NOT_FOUND`、`ALREADY_EXISTS`、`AMBIGUOUS_NOTE`、`PATH_OUTSIDE_VAULT`、`REVISION_CONFLICT`、`INVALID_FRONTMATTER`、`INVALID_ARGUMENT`、`PARTIAL_FAILURE`、`CAPABILITY_UNSUPPORTED`。

## 交付物

- `docs/spec/CLI_PROTOCOL.md`
- `docs/spec/schema/response-v2.schema.json`
- `testdata/protocol/*.json`

## 验收标准

- [x] 所有 envelope 字段的必填、可选和 null 语义明确。
- [x] JSON 输出中不存在人类装饰性文本。
- [x] 相同错误场景始终返回相同错误码。
- [x] Schema 能验证成功、失败和 dry-run 样例。
- [x] 文档说明兼容性和版本升级策略。

## 验证

```bash
go test ./... -run 'Protocol|JSON|ErrorCode'
```

## 验证记录

- 2026-07-24：JSON Schema 在 AJV Draft 2020-12 strict mode 下编译通过。
- 2026-07-24：成功、失败和 dry-run 三个黄金样例均通过 Schema 校验。
- 2026-07-24：Schema 与全部样例通过 `jq` JSON 语法检查。
- 2026-07-24：stdout/stderr、退出码、通用参数、错误码和兼容性关键词检查通过。
- 说明：协议实现测试将在 P1-T04 落地；当前任务验收的是规范与黄金数据。
