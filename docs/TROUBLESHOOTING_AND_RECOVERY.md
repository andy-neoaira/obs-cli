# 故障排查、备份与恢复

## 首要原则

先停止写入，再保存证据。不要删除 lock、transaction journal、冲突副本或损坏配置；
不要在未读最新 revision 时重放失败命令。

## 快速分流

| 现象/错误 | 常见原因 | 安全处理 |
|---|---|---|
| `CLI_UNAVAILABLE` | CLI 缺失或路径错误 | 保持纯插件模式；检查 `cli.command` |
| `CLI_PROTOCOL_INCOMPATIBLE` | 不是 `obs-cli/v2` | 禁用 Adapter，升级/降级到兼容版本 |
| `CLI_VAULT_CONTRACT_INCOMPATIBLE` | 共同规范不是 v1 | 禁用 Adapter，不执行 CLI 写入 |
| `CAPABILITY_UNSUPPORTED` | operation 缺失 | 不猜旧命令；只禁用该高级场景 |
| `REVISION_CONFLICT` | Obsidian/Neovim/同步服务已改文件 | 重新 get、比较并重新 plan |
| `PARTIAL_FAILURE` | 多文件事务未完整回滚 | 保存 transaction ID，逐项执行 recovery steps |
| 配置 JSON 损坏 | 中断、人工编辑或磁盘问题 | 先复制原文件，再从备份恢复或重新注册 |
| 笔记“消失” | 同步延迟、移动或冲突副本 | 暂停写入，检查所有客户端和同步历史 |

## 配置损坏

CLI V2 配置位于操作系统用户配置目录的 `obs-cli/config-v2.json`。处理步骤：

1. 停止所有 CLI/Agent 写入。
2. 复制损坏文件及同目录 lock，记录时间和报错。
3. 不覆盖 Obsidian 官方 `obsidian.json`。
4. 从已知良好备份恢复 CLI 配置；没有备份时，在临时配置根验证后重新执行
   `vault add`/`vault set-default`。
5. 用 `vault list` 和 `vault get` 核对路径后再恢复 Agent。

## 冲突与 dirty buffer

- CLI 冲突：保留失败 envelope，重新 `note get`，把最新磁盘内容与原计划比较。
- Neovim dirty buffer：选择“保留 buffer”或打开三方视图；在完成手动合并前写入门禁
  保持关闭。
- 同步冲突副本：把原文件、冲突副本和 Agent `revision_before` 三者一同保存，不以
  文件时间戳单独决定胜者。

## 部分失败与事务恢复

1. 保存 `transaction_id`、`plan_hash`、响应和 stderr。
2. 按响应中的 `recovery_steps` 顺序处理，不重新执行整次 move。
3. 检查 source、target、所有 rewritten notes 和私有恢复目录。
4. 完成后运行 backlinks/search 审计，确认没有新增坏链接。
5. 将证据附到故障记录后再解除写入冻结。

## 同步延迟

Obsidian Sync、iCloud、Git 或其他同步工具不参与 CLI 协作锁。每种同步方案都应记录：

- 典型与最坏同步延迟；
- 离线编辑策略；
- 冲突副本命名和位置；
- 文件历史保留期；
- 手机端插件和文件格式限制。

在所有客户端看到同一 revision 前，不执行依赖“最新状态”的批量 Agent 修改。

## 备份基线

- 发布/升级前：Vault 快照、CLI 配置、插件 lockfile/commit、CLI 二进制。
- 高风险移动前：先审查 dry-run，确保同步队列为空。
- 定期恢复演练：在隔离目录恢复一份 Vault，并运行 search、backlinks、Daily Note
  和三入口 E2E。

恢复顺序是“内容 → 配置 → 客户端集成”。任何时候都可以先关闭 CLI Adapter，让
插件回到独立模式。
