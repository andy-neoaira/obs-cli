# P4-T06：联合发布、兼容矩阵与运维文档

- 状态：`进行中（发布文档与自动门禁已完成，等待 P4-T05 人工门禁）`
- 负责人：`Codex（自动化）/ 待分配（人工验收）`
- 涉及项目：`obs-cli`、`miniobsidian.nvim`
- 依赖：P4-T05、P0-T06

## 目标

完成两个独立产品的联合发布说明、版本兼容关系、故障排查和回滚文档。

## 实施步骤

1. 建立机器可读兼容矩阵：CLI、协议、Vault 规范、插件、Skills。
2. CLI 和插件分别独立版本化、独立发布。
3. 记录插件无 CLI、兼容 CLI、不兼容 CLI 三种行为。
4. 编写安装、升级、降级和禁用可选集成说明。
5. 编写冲突、部分失败、配置损坏和同步延迟排查手册。
6. 编写备份与恢复建议。
7. 汇总 E2E、Skill eval、许可证和发布检查结果。
8. 更新本执行计划所有完成状态并归档最终报告。

## 交付物

- [机器可读兼容矩阵](../../compatibility.json)
- [联合使用指南](../../JOINT_USAGE.md)
- [故障排查与恢复手册](../../TROUBLESHOOTING_AND_RECOVERY.md)
- [联合验收报告](../../releases/AGENT_FIRST_V2_ACCEPTANCE.md)
- `scripts/compatibility-check.sh`

## 验收标准

- [x] 插件可单独安装和使用。
- [x] 每组兼容版本可以通过 capability 自动判断。
- [x] 升级和降级不会要求修改真实 Vault 内容格式。
- [x] 用户能关闭 CLI 集成并恢复纯插件模式。
- [ ] 发布物许可证、测试和 E2E 报告齐全。
- [x] 所有阶段任务状态与实际验收一致。

## 验证

```bash
cd /Users/andy/github/obs-cli
make release-check
./scripts/run-skill-evals.sh
./scripts/run-three-client-e2e.sh

cd /Users/andy/github/miniobsidian.nvim
make ci PLENARY_DIR=/path/to/plenary.nvim
```

## 当前结论

兼容矩阵、运行时 capability 判定、联合指南、排障恢复与自动发布证据已完成。
`obs-cli` 与插件仍独立安装、独立版本化；插件在 CLI 缺失、禁用、协议不兼容或
Vault 共同规范不兼容时保持本地功能可用。

正式验收仍依赖 P4-T05 的 Obsidian 桌面端 smoke 与移动端同步观察。完成前：

- `docs/compatibility.json` 保持 `release-candidate`；
- 验收报告保持“有条件通过”；
- 本任务、P4-T05 与 P4 阶段均不标记完成；
- 不创建或推送正式 tag/release。

2026-07-25 自动验证：

- `make release-check`：通过，覆盖率 72.5%。
- `./scripts/run-three-client-e2e.sh`：6/6 通过。
- `miniobsidian.nvim make ci`：格式、lint、fixture 和完整测试通过。
