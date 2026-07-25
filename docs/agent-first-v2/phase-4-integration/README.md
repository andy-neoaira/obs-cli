# P4：可选协同与端到端验收

## 阶段目标

在不破坏插件独立性的前提下提供高级协同，并验证 Obsidian、Agent、Neovim 操作同一 Vault 的完整闭环。

## 进入条件

- P1、P2、P3 已完成。
- 两个项目支持同一版本的 Vault 共同规范。

## 任务进度

阶段进度：`4 / 6`

- [x] [P4-T01 miniobsidian 可选 CLI Adapter](./P4-T01-optional-cli-adapter.md)
- [x] [P4-T02 Neovim 安全移动与批量操作入口](./P4-T02-safe-move-ui.md)
- [x] [P4-T03 当前笔记与选区交给 Agent](./P4-T03-agent-handoff.md)
- [x] [P4-T04 Agent 修改后的 Diff 与冲突 UX](./P4-T04-diff-conflict-ux.md)
- [ ] [P4-T05 三入口端到端测试](./P4-T05-three-client-e2e.md)（自动化 6/6；待桌面端/移动端人工冒烟）
- [ ] [P4-T06 联合发布、兼容矩阵与运维文档](./P4-T06-release-docs.md)（自动发布材料完成；待 P4-T05 人工门禁）

推荐顺序：P4-T01 → P4-T02/P4-T03 → P4-T04 → P4-T05 → P4-T06。

## 阶段完成标准

- [ ] 无 CLI 时插件功能与 P2 完成态一致。
- [ ] 有 CLI 时插件只通过 capabilities 决定高级功能。
- [x] Neovim 未保存内容不会被 Agent 覆盖。
- [ ] 三入口 E2E 覆盖同步、冲突、移动和 Daily Note。
- [x] 用户能清晰判断每次 Agent 修改的目标、差异和结果。
- [ ] 联合版本兼容关系有机器可读记录。

## 阶段验证

```bash
cd /Users/andy/github/obs-cli
./scripts/run-three-client-e2e.sh

cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/integration" -c qa
```
