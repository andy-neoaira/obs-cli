---
name: obsidian-capture
description: 将用户明确提供的想法、摘要、任务或 AI 输出安全捕获到 Obsidian 笔记。用户要求新建 Inbox 笔记或追加既有笔记时使用，支持 revision 冲突保护和写后验证。
---

# Obsidian Capture

## 触发条件

- 用户要求“记到 Obsidian”“保存到笔记”或“放到 Inbox”。
- 用户要求新建一篇捕获笔记。
- 用户要求把内容追加到明确的既有笔记。

## 非触发条件

- 只搜索、阅读、比较或总结 Vault 时不使用。
- 整理、移动或覆盖既有笔记时交给对应场景 Skill。
- 用户没有表达写入意图时，不因生成了摘要而自动保存。

## 输入

- `content`：必填，保留用户确认的 Markdown。
- `vault`：目标 Vault ID 或名称；多 Vault 或不明确时必须确认。
- `target_path`：Vault 相对路径；未指定时仅可在用户接受 Inbox 约定后推导。
- `mode`：`create` 或 `append`，必须明确。
- `section`：append 时可选的精确 ATX 标题。
- 多段内容默认按用户提供的原始顺序合并为一次 payload；段间换行不明确时先确认，不自行添加标题或分隔符。

默认建议为新建 `Inbox/<title>.md`；追加流水账可建议 `Inbox.md`。这些只是 Skill 约定，不是 obs-cli 或 Obsidian 特殊目录。

## Capability 前置检查

按模式检查完整 capability：

```bash
obs-cli capabilities --output json --require vault.get \
  --require note.get --require note.create
obs-cli capabilities --output json --require vault.get \
  --require note.get --require note.append
```

缺失时停止，报告缺失 operation 和当前 CLI 版本，并提示升级到支持对应 obs-cli/v2 operation 的版本。不得回退到 V1 `create --append`。

## 读取范围

- 只解析用户选择的 Vault。
- 只读取目标笔记；create 以结构化 `NOT_FOUND` 作为可创建前置条件。
- append 必须通过 `note get` 取得当前 revision 和必要上下文。

```bash
obs-cli vault get "<vault>" --output json
obs-cli note get "<target-path>" --vault "<vault>" --output json
```

不得为了捕获扫描整个 Vault，也不得把无关正文放入结果摘要。

## 写入范围

- create 只允许创建一个不存在的目标，不允许 overwrite。
- append 只允许向一个既有目标末尾或明确 section 追加。
- 禁止 replace、delete、move、修改 `.obsidian/` 或隐式创建其他文件。

## 执行生命周期

严格执行 `discover → read → plan → authorize → apply → verify`：

1. `discover`：检查 capability，解析唯一 Vault、路径和模式。
2. `read`：读取目标；append 记录 revision，create 确认目标不存在。
3. `plan`：生成稳定 request ID，将同一字节 payload 通过 stdin 或受控文件传入 `--dry-run`，检查目标和 diff。
4. `authorize`：确认计划未改变 Vault、路径、模式、section 或内容。
5. `apply`：使用相同内容和 request ID 执行一次；append 携带读取到的 `--if-match`。
6. `verify`：重新 `note get`，确认新 revision、内容摘要以及 create/append 不变量。

```bash
# 调用工具时把正文作为 stdin 传入，勿插入命令字符串。
obs-cli note create "<target-path>" --vault "<vault>" \
  --content-file - --dry-run --request-id "<stable-id>" --output json
obs-cli note create "<target-path>" --vault "<vault>" \
  --content-file - --request-id "<stable-id>" --output json

obs-cli note append "<target-path>" --vault "<vault>" \
  --content-file - --if-match "<revision>" --dry-run \
  --request-id "<stable-id>" --output json
obs-cli note append "<target-path>" --vault "<vault>" \
  --content-file - --if-match "<revision>" \
  --request-id "<stable-id>" --output json

obs-cli note get "<target-path>" --vault "<vault>" --output json
```

有 section 时在两次 append 命令中加入相同的 `--section "<exact-heading>"`。

## 授权策略

- 用户明确要求保存，且 Vault、路径、模式与内容无歧义时，dry-run 后可执行单笔低风险写入。
- 推导 Inbox 路径、改变 create/append 模式、改变 section 或扩大为多文件写入时先确认。
- 目标已存在时，create 立即停止；不得自动改成 append 或覆盖。

## 冲突、重试与幂等

- append 必须使用 `--if-match <revision>`。
- 遇到 `REVISION_CONFLICT` 时停止，重新读取并展示目标变化；不得移除前置条件重试。
- create 返回已存在时报告 `conflict`，不得使用 overwrite。
- 结果不明时先读取目标：create 内容与预期完全一致可返回 `no_change`；否则报告冲突。append 不自动重放，先比较写前/写后 revision 与尾部内容。
- 同一上层请求重投时复用 request ID；CLI 未提供持久幂等保证时，不得仅凭相同尾部文字推断重复请求。
- 多段内容合并为一次 append；verify 必须确认 payload 只新增一次、位于预期位置，且写前正文保持为写后正文的原样前缀。
- 所有参数使用结构化 argv 传递并经过路径/枚举校验，禁止把 Vault、路径、section 或 revision 插值为 shell 命令。
- 长文本和用户原文仅通过 stdin 或受控文件传入，绝不拼接到 shell 命令。

## 结果摘要

```yaml
status: success | no_change | conflict | partial | failed
vault:
  id: <resolved-id>
  name: <resolved-name>
path: <vault-relative-path>
operation: create | append
previous_revision: <revision-or-null>
new_revision: <verified-revision-or-null>
planned: [<dry-run summary>]
applied: [<actual change>]
verified:
  content_digest_matches: true | false
  mode_invariant_holds: true | false
conflicts: []
warnings: []
next_actions: []
```
