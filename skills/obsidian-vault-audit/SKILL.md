---
name: obsidian-vault-audit
description: 对 Obsidian Vault 执行有界只读审计，检查 TODO、坏链接、孤立笔记、重复标题和无效 frontmatter。用户要求盘点或生成整理报告时使用，只输出 revision 证据与后续修复计划。
---

# Obsidian Vault Audit

## 触发条件

- 用户要求检查 Vault 结构、TODO、坏链接、孤立笔记、重复标题或 frontmatter。
- 用户要求生成只读整理建议或健康报告。

## 非触发条件

- 不自动修复、移动、删除、重命名或修改 metadata。
- 不扫描附件内容、隐藏目录或 `.obsidian/`。
- 单一主题知识检索优先使用 Knowledge Search。

## 输入

- `vault` 必须唯一解析，`scope` 可选。
- `audit_types` 可从 `todo`、`broken_link`、`orphan`、`duplicate_title`、`frontmatter` 选择。
- 默认最多列举 1000 个路径、全文读取 100 篇；扩大前确认。
- 每类最多报告 100 条 finding，超出必须标记 truncated。

## Capability 前置检查

按审计类型要求：

```bash
obs-cli capabilities --output json --require vault.get
obs-cli capabilities --output json --require note.list --require note.get
obs-cli capabilities --output json --require search.content
obs-cli capabilities --output json --require link.backlinks
```

按 audit type 分别检查依赖。缺失 operation 时只将依赖类型标为 failed；其他类型仍可信时整体返回 `partial`，不得调用 V1 命令。

## 读取范围

- `note list` 只返回 Markdown 路径，因此附件和二进制文件不会被解析。
- TODO 使用 `search content`，每词最多 2 页、每页 50、`--max-files 1000`。
- duplicate title 明确定义为大小写敏感、Unicode 原值的 duplicate basename；最终候选必须 `note get` 取得 revision。
- frontmatter、outbound link 与 orphan 检查最多读取 100 篇；每条证据记录 path/revision/line/snippet。
- `INVALID_FRONTMATTER` 仅指 YAML 语法无效；details.path/revision 是有效 finding，line 可为 0。
- 100 篇全文是所有审计类型共享预算，不是每类型各 100。

## 写入范围

无。禁止任何 mutating operation、`--dry-run`、`--if-match` 和自动修复。修复建议只写入响应 `plan` 且 `execution_status=not_executed`；长文本仅通过 stdin 传递、不创建报告文件。后续执行必须交给相应修改型 Skill。

## 执行生命周期

只读执行 `discover → read → plan → authorize → apply → verify`：

1. `discover`：解析 Vault、scope、类型、限制和 capability。
2. `read`：list 后按类型有界搜索/get/backlinks；记录每次截断和错误。
3. `plan`：把候选归类为 finding，并生成不执行的修复 plan。
4. `authorize`：扩大 1000 路径或 100 篇全文限制前确认。
5. `apply`：仅在内存中 render report，不修改 Vault、不执行 plan。
6. `verify`：重验每个最终 positive finding 的 revision；变化的证据移除或重读。负检查在报告前立即重跑并记录 operation、target、scope、时间、分页耗尽与截断状态。

坏链接从已读取 Markdown 的 Wikilink/Markdown link 目标解析；外部 URL 和纯锚点跳过。`.pdf`、图片等附件目标在没有附件枚举 capability 时标为 `unverified_attachment_target` 并令该类型 partial，绝不报 broken_link。Markdown 目标必须通过直接 `note get` 得到 `NOTE_NOT_FOUND` 才能报告 observed broken_link；list 截断不能作为不存在证据。

当前 capability 没有 Vault-wide snapshot，其他入口可能在检查后新增链接，因此零 backlinks 始终只报告 `orphan_candidate`，整体至少 partial。必须在输出前重跑完整、无错误、未截断的 backlinks，并记录负检查；任何失败都不得作为“无入站链接”的证据。

## 授权策略

- 默认只读范围无需额外确认。
- 大 Vault 超过限制时返回 `partial`，不得自动扩大扫描。每个类型单独记录 eligible、processed、status 与 stop_reason。
- 用户要求修复时仅输出 plan，并指出建议使用的 Skill；必须另行授权执行。

## 冲突、重试与幂等

- 本 Skill 不使用 `--if-match` 或 `--dry-run`。
- 证据 revision 改变或出现 `REVISION_CONFLICT` 时重读一次；持续变化则移出确定 finding 并返回警告。
- 无结果是成功的零 finding；命令失败记录到 errors，不能伪装为“没有问题”。
- 只有相关分页耗尽且 coverage=complete 时才可报告该类型零问题。
- 重试保持相同 scope、类型和上限，报告排序按 category/path/line 稳定。
- 任何失败都不会触发写入或修复。

## 结果摘要

遵循 `docs/spec/schema/search-audit-report-v1.schema.json`。每个 finding 通过 `evidence_indexes` 指向 path、revision、line、snippet；修复建议只进入：

```yaml
plan:
  - action: repair_broken_link
    targets: [Source.md]
    requires_skill: obsidian-safe-note-update
    execution_status: not_executed
checks:
  - operation: note.get
    target: Missing.md
    scope: ""
    observed_at: "2026-07-25T12:00:00+08:00"
    status: absent
    pagination_complete: true
    scan_truncated: false
coverage:
  - type: broken_link
    status: partial
    eligible: 250
    processed: 100
    scan_truncated: true
    findings_truncated: false
    stop_reason: shared_full_read_budget
pagination:
  pages_read: 2
  files_scanned: 1000
  full_notes_read: 100
  truncated: true
  stop_reason: audit_read_limit
```
