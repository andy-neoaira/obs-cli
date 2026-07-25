---
name: obsidian-project-note
description: 创建或增量更新项目级 Obsidian 笔记及其 frontmatter。用户要求保存项目分析、决策、TODO、架构或 session log 时使用，避免整篇覆盖并保护未知 metadata。
---

# Obsidian Project Note

## 触发条件

- 用户要求创建项目 Overview、Architecture、Decisions、TODO 或 Session Log。
- 用户要求向明确项目笔记的章节追加内容。
- 用户要求更新 `project`、`status`、`type` 或 `note_type` 字段。

## 非触发条件

- 只分析项目但未要求写入时不保存。
- 不用于 Daily Note、Inbox 整理或跨文件批量重构。
- 不执行整篇 replace、delete 或隐式 move。

## 输入

- `project_name`、`vault`、`target_path` 和写入意图必须明确。
- `content` 可选；有正文时指定唯一 `section` 或明确 create。
- `metadata` 为用户要求修改的键值集合；不推断无关字段。

可建议 `Projects/<project_name>/<NoteType>.md`，但推导路径必须确认。

## Capability 前置检查

按实际分支检查，不要求不会使用的 create：

```bash
obs-cli capabilities --output json --require note.get
obs-cli capabilities --output json --require note.append
obs-cli capabilities --output json --require note.create
obs-cli capabilities --output json --require metadata.get \
  --require metadata.set
```

缺失 capability 时停止并给出升级提示；不得使用 V1 `create --append`、`--overwrite` 或 `frontmatter --edit`。

## 读取范围

- 只读取目标项目笔记。
- `note get` 获取正文、section 和 revision；`metadata get` 必须返回相同 revision，否则重新读取两者，禁止用撕裂快照规划。
- 不扫描整个 Projects 目录，除非用户另行授权搜索。

## 写入范围

- 不存在的目标可用 `note create` 创建一次。
- 已有目标仅允许 `note append --section` 或 `metadata set`。
- 每次 metadata set 只更新一个明确字段，保留未知字段和正文。
- 禁止整篇覆盖、删除、移动或修改其他项目笔记。

## 执行生命周期

执行 `discover → read → plan → authorize → apply → verify`：

1. `discover`：检查 capability，解析唯一 Vault 与路径。
2. `read`：读取 note 和 metadata，记录 revision、目标 section 与未知字段。
3. `plan`：展示组合操作总览，但每个实际写入步骤都必须基于最新 revision 单独 dry-run。
4. `authorize`：核对路径、section、字段和值；批量字段展示完整计划。
5. `apply`：每步在 apply 前重新读取；状态变化则重做该步 dry-run/必要授权。正文写入后验证并取得新 revision，再规划 metadata。
6. `verify`：apply 返回 revision 必须与紧随其后的读取一致；若出现后继外部修改则单独报告。确认 section delta、目标字段、所有非目标字段和值及正文不变量。

```bash
obs-cli note get "<path>" --vault "<vault>" --output json
obs-cli metadata get "<path>" --vault "<vault>" --output json

obs-cli note create "<path>" --vault "<vault>" --content-file - \
  --dry-run --request-id "<stable-id>" --output json
obs-cli note append "<path>" --vault "<vault>" --content-file - \
  --section "<exact-heading>" --if-match "<revision>" \
  --dry-run --request-id "<stable-id>" --output json

obs-cli metadata set "<path>" --vault "<vault>" \
  --key "<key>" --value "<value>" --if-match "<revision>" \
  --dry-run --request-id "<field-id>" --output json
```

apply 使用该步骤刚通过的相同业务参数去掉 `--dry-run`；每步之后重新 get，再 dry-run/执行下一步骤。

## 授权策略

- 用户明确要求写入，且 Vault、路径、section、字段和值无歧义时，可在 dry-run 一致后执行。
- 推导路径、新建笔记、新建缺失 section、多个字段或计划范围扩大时先展示并确认。
- 不把“项目分析”解释为保存授权。

## 冲突、重试与幂等

- 所有既有笔记修改必须使用 `--if-match <revision>`。
- `REVISION_CONFLICT` 后停止剩余字段，重新读取并重新规划；不得强制覆盖。
- create 冲突时读取现有笔记，不自动改成 append。
- section 已有相同 payload 或 metadata 已是目标值时返回 `no_change`。
- 重复 payload 只在同一请求恢复时依据写前/写后 section delta 和持久 request ID 判断；普通相同文本不自动去重。
- combined 任一步后失败都返回 `partial`，列出已验证写入、失败步骤、当前 revision 和未执行项。
- 不自动回滚已验证的正文或 metadata；重入时验证并跳过已完成步骤。
- metadata 保证所有非目标键的值与 YAML 类型语义不变；不承诺保留 YAML 键顺序、引号样式等纯格式。
- section 使用去掉 `#` 的精确 ATX 标题文本；0 个需确认创建，多个匹配停止。
- 所有参数通过结构化 argv；正文只通过 stdin 或受控文件。

## 结果摘要

```yaml
status: success | no_change | conflict | partial | failed
vault: <resolved-vault>
path: <project-note-path>
operation: create | append-section | set-metadata | combined
previous_revision: <revision-or-null>
new_revision: <verified-revision>
metadata:
  changed: {}
  preserved_unknown: []
section: <heading-or-null>
planned: []
applied: []
verified:
  body_preserved: true
  unknown_metadata_preserved: true
  section_payload_added_once: true
conflicts: []
warnings: []
next_actions: []
```
