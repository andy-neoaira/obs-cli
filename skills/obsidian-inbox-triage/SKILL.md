---
name: obsidian-inbox-triage
description: 安全整理 Obsidian Inbox，将明确的源笔记事务化移动到已确认目标、重写入站链接，并可恢复地更新 metadata。用户要求分类、归档或移动 Inbox 笔记时使用。
---

# Obsidian Inbox Triage

## 触发条件

- 用户要求整理 Inbox、分类临时笔记、归档或移动到正式目录。
- 用户要求移动后更新 status/classification 等明确 metadata。

## 非触发条件

- 只查看、搜索或审计 Inbox 时使用只读 Skill。
- 不用于删除笔记、整篇覆盖或自动执行分类建议。
- 不打开编辑器或 Obsidian GUI。

## 输入

- `vault` 与 `source_path` 必须唯一。
- `target_path`、`classification` 和 metadata 可由 Agent 建议，但任何推断目标必须确认。
- 批次默认最多 20 篇；每篇记录 source revision、目标存在状态和用户授权。
- `Inbox` 只是 Skill 约定，不是 CLI 特殊目录。
- CLI 只能保护 Vault 磁盘快照，不能感知 Obsidian/Neovim 的未保存缓冲区，也不能强制外部应用遵守 CLI 协作锁。

## Capability 前置检查

按实际步骤检查：

```bash
obs-cli capabilities --output json --require vault.get \
  --require note.list --require note.get --require note.move
obs-cli capabilities --output json --require metadata.get \
  --require metadata.set --require link.backlinks
```

没有 metadata 请求时不要求 metadata operation。缺失能力时停止相关步骤，不回退到 V1 `move`、`frontmatter --edit` 或文件系统命令。

## 读取范围

- `note list` 后只选择 `Inbox/` 下最多 20 个明确 source。
- 每篇用 `note get` 记录 content digest、frontmatter 和 revision。
- 对每个 target 执行 `note get`：只有 `NOTE_NOT_FOUND` 才可计划 move。
- 用 `link backlinks` 记录 move 前入站链接及 source revision，最多扫描 1000 个 Markdown 文件；若报告 truncated，则该篇只能返回 `partial`。
- dry-run 后记录 `plan_hash`；它绑定 source、target、全部链接改写、各文件 revision 和风险。

## 写入范围

- 每篇只允许一次 `note move <source> <target>`；该 operation 原子提交目标创建、相关链接重写和源删除。
- move 成功后才允许在目标上逐字段执行 `metadata set`。
- metadata 是后置可恢复步骤，不属于 move 的多文件事务；不得声称两者整体原子。
- 禁止覆盖已存在目标、delete、replace、修改 `.obsidian/` 或写入未授权笔记。

## 执行生命周期

执行 `discover → read → plan → authorize → apply → verify`：

1. `discover`：解析 Vault、Inbox 候选和 capability。
2. `read`：读取 source revision、目标不存在证据和 backlinks。
3. `plan`：为每篇生成 move dry-run；合并展示 source/target、链接改写、metadata 后置步骤、风险和 `plan_hash`。
4. `authorize`：逐篇确认所有推断 target；用户可批准子集。进入 apply 前建立静默窗口：要求保存或关闭 Obsidian/Neovim 中 source 与计划受影响笔记的未保存缓冲，并在 verify 完成前不要从其他入口编辑该 Vault。
5. `apply`：使用用户已授权的 `plan_hash` 执行。CLI 会重新规划；任何 source、target、backlink 集合或 revision 变化都会使 hash 不匹配并停止。先执行一次事务化 move，再逐字段重新 get/dry-run/apply metadata。
6. `verify`：立即保存 move receipt 到工作流结果；确认 source 为 `NOTE_NOT_FOUND`、target revision/body revision 正确；对已不存在的 source 再运行 `link backlinks`，结果必须完整且为空，才可确认旧链接已重写；验证每个 metadata 字段。

```bash
obs-cli note move "<source>" "<target>" --vault "<vault>" \
  --if-match "<source-revision>" --dry-run \
  --request-id "<move-plan-id>" --output json
obs-cli note move "<source>" "<target>" --vault "<vault>" \
  --if-match "<source-revision>" --if-plan-hash "<authorized-plan-hash>" \
  --request-id "<move-apply-id>" --output json
obs-cli link backlinks "<source>" --vault "<vault>" \
  --max-files 1000 --request-id "<verify-old-links-id>" --output json

obs-cli metadata set "<target>" --vault "<vault>" \
  --key status --value organized --if-match "<target-revision>" \
  --dry-run --request-id "<metadata-id>" --output json
```

metadata apply 使用刚验证的相同业务参数去掉 `--dry-run`，随后重新 get；多个字段串行取得新 revision。

## 授权策略

- list/get/backlinks/dry-run 可直接执行。
- 用户明确给出 source 和 target 时，可在 move plan 一致、没有未保存编辑器缓冲且用户接受 apply→verify 静默窗口的前提下执行。
- 推断的分类、target、重命名、metadata 值和批量子集必须展示并确认。
- target 已存在、同名路径歧义或计划出现未授权链接变化时停止，不覆盖、不自行换名。

## 冲突、重试与幂等

- move 必须同时使用已授权 dry-run 的 `--if-match` 与 `--if-plan-hash`；metadata 使用自己的最新 `--if-match`。
- 遇到 `REVISION_CONFLICT` 时停止相关篇目并重新规划，绝不强制覆盖。
- source、target、backlink 集合或任一受影响文件 revision 变化只停止相关篇目；重新读取、重新规划并重新授权，绝不移除前置条件。
- `note.move` 内部冲突必须保持 source/target/links 原状；若返回 partial，严格采用响应中的 rollback/recovery actions。
- `plan_hash` 保护 apply 开始时重新扫描到的计划和 CLI 协作写入，但外部应用不遵守锁；apply 后旧 source backlinks 非空或扫描 truncated 时返回 `partial/concurrent_external_change`，不得声称链接一致，也不得自动 patch 未授权的新笔记。
- move 已验证而 metadata 失败时返回 `partial`：目标和链接保持有效，列出 pending fields 与当前 target revision，不回滚 move。
- move 成功后必须原样保存响应 `receipt` 到当前工作流结果或 Agent task state；不得只保存自然语言摘要。receipt 至少绑定 operation、request/transaction ID、plan hash、Vault ID、source revision/digest、target revision 和 `target_body_revision`。
- 每个 metadata 字段成功后追加 receipt step：`key/value/revision_before/revision_after/body_revision/changed`。下一步只允许使用上一 step 的 `revision_after`，且 `body_revision` 必须始终等于 move receipt。
- 同一工作流恢复时，必须提供原 receipt，并用 `metadata get <target>` 校验 Vault/path、完整 revision、body revision 和已完成字段。当前 revision 必须等于最后一个 metadata step 的 `revision_after`（没有 step 时等于 move target revision），body revision 必须等于 move receipt；否则按外部编辑冲突停止并重新展示差异、请求授权，不能直接继续 pending metadata。
- source 不存在且 move receipt、旧链接复验和 metadata revision 链均可验证时才可跳过 move。所有字段已具备预期值且 revision 链未变化时返回 `no_change`；若仍执行同值 `metadata set`，CLI 的 `changed` 必须为 false。
- 没有原计划/receipt 时不得仅凭 source 不存在或 target 存在猜测 move 已完成，返回 `failed` 并请求人工确认。
- 重复执行已完成工作流返回 `no_change`，不得重复移动、重写链接或写入相同 metadata。
- request ID 仅用于关联，不替代 revision 或幂等验证；长文本只通过 stdin 或受控文件。

## 结果摘要

```yaml
status: success | no_change | conflict | partial | failed
vault: <resolved-vault>
items:
  - source: Inbox/tmp.md
    source_revision: "sha256:..."
    target: Projects/Alpha/Note.md
    target_revision: "sha256:..."
    move_status: applied | no_change | conflict | failed
    receipt:
      operation: note.move
      request_id: <move-apply-id>
      transaction_id: <transaction-id>
      plan_hash: "sha256:..."
      vault_id: <vault-id>
      source: Inbox/tmp.md
      source_revision: "sha256:..."
      source_digest: "sha256:..."
      target: Projects/Alpha/Note.md
      target_revision: "sha256:..."
      target_body_revision: "sha256:..."
    metadata_steps:
      - key: status
        value: organized
        revision_before: "sha256:..."
        revision_after: "sha256:..."
        body_revision: "sha256:..."
        changed: true
    link_rewrites: []
    metadata_applied: {}
    metadata_pending: {}
    verified:
      source_absent: true
      target_matches: true
      backlinks_consistent: true
      verification_complete: true
recovery_actions: []
warnings: []
```
