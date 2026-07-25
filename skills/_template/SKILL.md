---
name: replace-with-skill-name
description: 说明 Skill 完成什么任务，并在此列出应触发它的具体用户意图和上下文。
---

# Skill 标题

## 触发条件

- 列出代表性的用户意图和边界清晰的场景。

## 非触发条件

- 列出相邻但应交给其他 Skill 或只需普通回答的场景。

## 输入

- 区分必填、可选、可推断和必须向用户确认的输入。
- 不把未验证的推断当作写入目标。

## Capability 前置检查

执行任何 Vault 操作前运行：

```bash
obs-cli capabilities --output json --require <operation>
```

声明本 Skill 使用的只读与修改型 operation；缺失 capability 时停止，不回退到旧命令。

## 读取范围

- 声明允许读取的 Vault、目录、笔记和最大结果量。
- 先缩小候选集，再读取必要正文。

## 写入范围

- 只读 Skill 明确写“无”，且不得执行修改型 operation。
- 修改型 Skill 声明目标、允许的变更类型和禁止操作。

## 执行生命周期

严格按 `discover → read → plan → authorize → apply → verify` 执行：

1. `discover`：解析 Vault，检查 capability 和输入。
2. `read`：读取最小必要范围及目标当前 revision。
3. `plan`：生成 dry-run 计划，列出目标、diff、风险和前置条件。
4. `authorize`：按授权策略决定继续、询问或停止。
5. `apply`：携带 dry-run 使用的目标与 revision 执行一次写入。
6. `verify`：重新读取目标，验证 revision、内容和不变量。

只读 Skill 执行 `discover → read → plan → verify`，并在 plan 中明确无写入。

## 授权策略

- 用户原始请求已明确授权且目标、范围、变更方式均无歧义时，可执行低风险写入。
- 覆盖、删除、批量移动、扩大读取/写入范围或改变默认 Vault 前必须确认。
- dry-run 与用户授权不一致时重新授权，不得自行扩大范围。

## 冲突、重试与幂等

- 更新与删除必须传 `--if-match <revision>`。
- 遇到 `REVISION_CONFLICT` 时停止，重新读取并展示差异；不得无条件重试或覆盖。
- 使用稳定目标和 request ID；重试前确认上次操作是否已生效。
- 部分失败时逐项汇报成功、失败和未执行项，不把整批报告为成功。
- 长文本仅通过 stdin 或受控文件传入，不拼接进 shell 命令。

## 结果摘要

返回结构化摘要：

```yaml
status: success | no_change | conflict | partial | failed
vault: <resolved-vault>
read:
  - <path@revision>
planned:
  - <dry-run change>
applied:
  - <actual change>
verified:
  - <post-write check>
conflicts: []
warnings: []
next_actions: []
```
