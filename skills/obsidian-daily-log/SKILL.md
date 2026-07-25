---
name: obsidian-daily-log
description: 将工作日志、会议记录、session 总结或待办安全追加到 Obsidian Daily Note。用户明确要求记录到某日 Daily Note 时使用，遵循官方 Daily Notes 路径、模板与 revision。
---

# Obsidian Daily Log

## 触发条件

- 用户要求写入今天或明确日期的 Daily Note。
- 用户要求追加工作日志、会议摘要、TODO 或 session 总结。

## 非触发条件

- 只查看或总结日记时不执行写入。
- 不用于普通 Inbox 捕获或项目笔记维护。
- 不修改 `.obsidian/daily-notes.json`、模板或日期规则。

## 输入

- `content`：必填，按用户确认的 Markdown 字节传入。
- `vault`：目标 Vault；无法唯一解析时确认。
- `date`：可选 ISO 日期 `YYYY-MM-DD`，默认今天。
- `section`：可选精确 ATX 标题；不得根据相似标题猜测。

多段内容合并成一次 payload；段间格式不明确时先确认。

## Capability 前置检查

```bash
obs-cli capabilities --output json --require daily.get
obs-cli capabilities --output json --require daily.append
obs-cli capabilities --output json --require daily.create
```

先检查 `daily.get`；确认目标存在后只要求 append，不存在时再要求 create 与 append。实际分支缺失 operation 时停止并提示升级；不得回退到 V1 `obs-cli daily`。

## 读取范围

- `daily get` 读取官方 Daily 配置解析出的单个目标及 revision。
- 不硬编码 `Dailies/`、日期格式或模板路径。
- 创建后、追加前和写后均重新 get。

```bash
obs-cli daily get --date "<YYYY-MM-DD>" --vault "<vault>" --output json
```

## 写入范围

- 目标不存在时仅允许 `daily create` 按官方配置和模板创建。
- 目标存在时不得重复 create。
- 正文只通过 `daily append` 追加到目标末尾或唯一 section。
- 禁止覆盖整篇、修改模板、修改官方配置或写入其他日期。

## 执行生命周期

执行 `discover → read → plan → authorize → apply → verify`：

1. `discover`：检查 get capability，以 Vault 所在系统时区把“今天”冻结为一个 ISO 日期；同一请求跨午夜不改变日期。
2. `read`：运行 `daily get`，取得规范路径、存在状态和 revision。
3. `plan`：若不存在，先 dry-run create；存在后用当前 revision dry-run append。
4. `authorize`：核对 Vault、日期、路径、section 和 payload。
5. `apply`：apply 前再次 get；路径、revision 或 section 状态变化则废弃旧计划，重新 dry-run/授权。随后执行一次 create/append。
6. `verify`：记录 apply 返回 revision 并重新 get；用写前/写后结构化 delta 验证目标日期、section 和精确新增内容。

```bash
obs-cli daily create --date "<YYYY-MM-DD>" --vault "<vault>" \
  --dry-run --request-id "<create-id>" --output json
obs-cli daily create --date "<YYYY-MM-DD>" --vault "<vault>" \
  --request-id "<create-id>" --output json

obs-cli daily append --date "<YYYY-MM-DD>" --vault "<vault>" \
  --content-file - --section "<exact-heading>" --if-match "<revision>" \
  --dry-run --request-id "<append-id>" --output json
obs-cli daily append --date "<YYYY-MM-DD>" --vault "<vault>" \
  --content-file - --section "<exact-heading>" --if-match "<revision>" \
  --request-id "<append-id>" --output json
```

无 section 时省略两条 append 命令中的 `--section`。

## 授权策略

- 用户明确要求写入指定日期且目标解析唯一时，可在 dry-run 一致后执行。
- 推导 Vault、改变日期、创建缺失 section 或改变内容格式时先确认。
- section 参数是去掉 `#` 的精确 ATX 标题文本；0 个匹配时停止确认，多个匹配时报歧义。用户确认后可将“标题 + 内容”作为一次无 section 文末追加，但不得声称写入了既有章节。再次执行前先检查标题，禁止重复创建章节。

## 冲突、重试与幂等

- append 必须携带 `--if-match <revision>`。
- `REVISION_CONFLICT` 后重新 get 并展示变化，不移除前置条件重试。
- create 返回已存在时重新 get，若同一 Daily 已由 Obsidian/Neovim 创建则继续读取，不覆盖。
- create 已完成但 append 失败时返回 `partial`，保留已创建日记，不回滚或重建。
- 结果不明时先验证路径、revision 和 payload；不得自动重复追加。
- request ID 仅用于关联，不视为服务端幂等键；同一上层请求需持久复用，独立的新请求不因文本相同而自动去重。
- `payload_added_once` 通过写前/写后 section delta 验证，不通过全文出现次数猜测。
- verify 读到 apply revision 的后继版本且 payload 仍保持时，报告外部后续修改警告；无法证明本次 delta 时不得返回 success。
- section 标题存在时只向该节追加，不再创建同名章节。
- 所有参数使用结构化 argv；长文本只通过 stdin 或受控文件。

## 结果摘要

```yaml
status: success | no_change | conflict | partial | failed
vault: <resolved-vault>
date: <YYYY-MM-DD>
path: <resolved-daily-path>
operation: create+append | append
previous_revision: <revision-or-null>
new_revision: <verified-revision>
section: <heading-or-null>
planned: []
applied: []
verified:
  official_path_matches: true
  section_unique: true
  payload_added_once: true
conflicts: []
warnings: []
next_actions: []
```
