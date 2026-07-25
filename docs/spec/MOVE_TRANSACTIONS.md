# Note Move 与链接重写事务

- CLI operation：`note.move`
- 写入协议：`obs-write/v1`
- Vault 规范：`vault-contract/v1` VC-5
- 状态：P1-T07 已实现

## 1. 调用

```bash
obs note move <source> <target> \
  --vault <id-or-name> \
  --if-match sha256:<64-lowercase-hex> \
  --dry-run

obs note move <source> <target> \
  --vault <id-or-name> \
  --if-match sha256:<64-lowercase-hex> \
  --if-plan-hash sha256:<64-lowercase-hex>
```

source/target 是 Vault 逻辑路径。目标默认补充 `.md`，已存在时返回 `ALREADY_EXISTS`，绝不依赖可能覆盖目标的平台 rename 行为。

## 2. Plan

plan 在首次写入前完成：

1. 解析 source/target 的规范物理身份并拒绝写入符号链接别名。
2. 验证 source revision 和 target 不存在。
3. 读取全部候选 Markdown 的稳定快照。
4. 解析有效 Wikilink 与 Markdown link，计算每个文件的精确 link edits。
5. 记录 move、所有 rewrite、revision before/after 和风险。

`--dry-run` 返回该 plan 及确定性的 `plan_hash`，不创建目标目录、锁、临时文件、
journal 或恢复副本。`plan_hash` 绑定协议版本、source、target、全部 changes 和风险。
Agent 应展示 dry-run 后保存该 hash，并在 apply 时传入 `--if-plan-hash`。

## 3. 重写边界

- Wikilink 分离 target、heading/block fragment 与 alias，只修改 target。
- `[[target|alias]]`、`[[target#heading|alias]]` 保留 alias 与 fragment。
- basename 只有在 Vault 中唯一时才会被重写；同名歧义时只重写完整 Note ID。
- Markdown link 按包含链接的 Note 目录解析相对路径，安全 percent-decode 一次，再生成指向新目标的相对路径。
- URL、外部 scheme 和 Vault 绝对风格目标不参与重写。
- Frontmatter、fenced code、inline code、HTML comment 和普通文本中的相似字符串不参与重写。

## 4. Apply 与恢复

apply 将 target create、受影响文件 update 和 source delete 作为一个多文件事务：

- 按稳定路径顺序获取协作锁；
- stage 前验证全部 revision/target-not-exist 前置条件；
- 在受控运行时目录保存原内容、权限和 transaction journal；
- 任一普通提交失败时逆序回滚；
- plan 后发生外部修改时返回 `REVISION_CONFLICT`，不开始部分提交；
- `--if-plan-hash` 与重新计算的 plan 不一致时返回 `REVISION_CONFLICT`，因此用户
  未授权的新 backlink 或 revision 变化不会被静默纳入 apply；
- 回滚不完整时返回 `PARTIAL_FAILURE`。

成功响应包含结构化 `receipt`，用于工作流恢复和去重判断：

- `operation`、`request_id`、`transaction_id`
- `plan_hash`、`vault_id`
- `source`、`source_revision`、`source_digest`
- `target`、`target_revision`、`target_body_revision`

receipt 必须由 Agent 保存到工作流结果或 task state；request ID 本身不是幂等键。

CLI 的协作锁不能阻止 Obsidian/Neovim 直接写文件。跨入口工作流必须建立短暂静默窗口，
并在 apply 后对已不存在的 source 调用 `link.backlinks`：该只读操作允许 target 缺失，
通过 `target_exists: false` 和完整、未截断的空结果证明没有遗留旧链接。非空或 truncated
结果只能报告 partial，不能声称 move 与外部编辑整体原子。

`PARTIAL_FAILURE.details` 包含：

- `transaction_id`
- `completed[]`
- `failed[]`
- `rolled_back[]`
- `rollback_failed[]`
- `recovery_actions[]`

路径均为 Vault 逻辑路径，不暴露运行时恢复目录或 Vault 外绝对路径。调用方必须先处理恢复清单，不能直接重放整个 move。
