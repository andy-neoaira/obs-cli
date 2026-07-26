# Agent-first V2 联合验收报告

- 报告日期：2026-07-25
- 状态：`有条件通过（自动化与桌面端完成，移动端门禁待执行）`
- 产品：`obs-cli`、`miniobsidian.nvim`、Agent Skills
- 兼容事实源：[compatibility.json](../compatibility.json)

## 已验证

- `obs-cli make release-check`：format、vet、test、race、覆盖率、Schema、六目标构建、
  license、Skill lint/eval 与 RC smoke 全部通过。
- `miniobsidian.nvim make ci`：StyLua、Selene、固定 fixture 与完整 Plenary
  回归通过。
- `scripts/run-three-client-e2e.sh`：六个合成三入口场景通过且黄金摘要一致。
- CLI capabilities 与插件 Adapter 同时检查 `obs-cli/v2`、
  `vault-contract/v1` 和所需 operation。
- 两项目许可证边界、派生来源和发布归档要求已记录。
- 插件在 CLI 缺失、关闭或不兼容时保留全部本地核心功能。
- Obsidian Desktop `1.12.7` 在专用临时 Vault 完成真实 UI smoke：
  CLI/Neovim 读写闭环、revision 冲突保护、移动改链和 Daily Note 均通过。

## 未完成门禁

- 移动端同步延迟、离线冲突和冲突副本观察记录。

执行清单见
[P4-T05 人工 Obsidian 冒烟清单](../agent-first-v2/phase-4-integration/P4-T05-manual-smoke.md)。
在移动端门禁完成前，不创建正式联合发布结论，不把
P4-T05/P4-T06/P4 阶段标为完成。

## 发布决策

当前代码满足 release candidate 的自动化质量门禁，可以进入人工验收；不满足正式
稳定发布门禁。两个项目继续独立版本化和发布，兼容性由机器矩阵与 runtime
capabilities 决定，不通过共享包或强制依赖耦合。

## 回滚

1. 关闭 `miniobsidian.nvim` 的 CLI Adapter。
2. 停止 Agent 写入并保存失败 envelope/transaction evidence。
3. 回退 CLI 二进制或插件 commit。
4. 用同步历史、版本控制或快照恢复业务内容；二进制回退不自动撤销 Markdown 修改。
5. 在隔离 Vault 重跑门禁后再恢复集成。
