# P4-T06：联合发布、兼容矩阵与运维文档

- 状态：`未开始`
- 负责人：`待分配`
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

- `docs/compatibility.json`
- 联合使用指南
- 故障排查与恢复手册
- 最终验收报告

## 验收标准

- [ ] 插件可单独安装和使用。
- [ ] 每组兼容版本可以通过 capability 自动判断。
- [ ] 升级和降级不会要求修改真实 Vault 内容格式。
- [ ] 用户能关闭 CLI 集成并恢复纯插件模式。
- [ ] 发布物许可证、测试和 E2E 报告齐全。
- [ ] 所有阶段任务状态与实际验收一致。

## 验证

```bash
cd /Users/andy/github/obs-cli
make release-check
./scripts/run-skill-evals.sh
./scripts/run-three-client-e2e.sh

cd /Users/andy/github/miniobsidian.nvim
stylua --check .
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests" -c qa
```
