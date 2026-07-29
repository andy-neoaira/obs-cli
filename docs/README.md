# obs-cli 当前文档

本目录只保存当前 V1 的产品说明、架构决策、规范、机器契约和持续有效的合规信息。
一次性重构计划、完成清单、旧兼容映射以及绑定历史 commit 的候选验证快照不在当前
文档树中归档；需要追溯时使用 Git 历史。

## 用户与集成

- [命令参考](./COMMAND_REFERENCE.md)
- [obs-cli 与 miniobsidian.nvim 联合使用](./JOINT_USAGE.md)
- [故障排查、备份与恢复](./TROUBLESHOOTING_AND_RECOVERY.md)
- [兼容矩阵](./compatibility.json)

## 架构与规范

- [Agent-first 产品边界](./architecture/ADR-001-agent-first-boundary.md)
- [CLI JSON 协议](./spec/CLI_PROTOCOL.md)
- [Capability 与 dry-run](./spec/CAPABILITIES.md)
- [配置](./spec/CONFIG.md)
- [Vault 约定](./spec/VAULT_CONVENTIONS.md)
- [路径策略](./spec/PATH_POLICY.md)
- [Note 操作](./spec/NOTE_OPERATIONS.md)
- [Move 事务](./spec/MOVE_TRANSACTIONS.md)
- [Revision、原子写入与冲突](./spec/CONCURRENCY_AND_WRITES.md)
- [原子存储实现](./spec/ATOMIC_STORAGE.md)
- [Skill 契约](./spec/SKILL_CONTRACT.md)
- [Agent Handoff](./spec/AGENT_HANDOFF.md)
- [Agent Result](./spec/AGENT_RESULT.md)

`spec/schema/` 中的 JSON Schema 是 CI 校验和跨客户端消费的机器契约，不是历史文档。

## 合规

- [开源许可与来源审查](./legal/LICENSE_REVIEW.md)
- 根目录 [`LICENSE`](../LICENSE) 与
  [`THIRD_PARTY_NOTICES.md`](../THIRD_PARTY_NOTICES.md)

## 文档收录规则

新增文档必须至少满足一项：

1. 描述当前用户可用行为；
2. 约束当前运行时或跨项目接口；
3. 被 CI、Schema、Skill 或发布流程直接消费；
4. 保存持续有效的架构或合规决策。

执行过程、临时验证输出、已完成迁移步骤和废弃接口映射应留在 commit、PR 或 release
记录中，不再复制到当前文档树。

本地运行 `make docs-check` 可检查当前 Markdown 相对链接，并阻止已退休文档名称重新
进入当前说明、脚本和 Skill。
