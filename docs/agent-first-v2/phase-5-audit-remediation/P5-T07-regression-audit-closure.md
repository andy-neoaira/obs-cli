# P5-T07：全量回归与审计关闭报告

- 状态：`待开始`
- 优先级：`高`
- 负责人：`待分配`
- 涉及项目：`obs-cli`、`miniobsidian.nvim`、Agent Skills
- 依赖：P5-T01 至 P5-T06
- 来源：P5 阶段收口

## 目标

对六项修复执行联合回归，形成逐项可追踪的审计关闭报告并更新任务事实源。

## 非目标

- 不在回归阶段修复新发现的问题；新问题单独建任务。
- 不执行真实移动端同步门禁。
- 不把兼容矩阵从 `release-candidate` 改为 `stable`。
- 不创建或推送 tag/release。

## 交付物

- `docs/releases/P5_AUDIT_REMEDIATION.md`
- P5 各任务完成记录
- P5 阶段 `7 / 7` 状态
- 根执行计划 P5 `[x]`

## 执行过程

1. 确认 T01-T06 每项都有独立提交、失败测试和完成记录。
2. 运行 obs-cli 完整发布门禁。
3. 运行三入口 6/6 自动 E2E，验证删除 V1 和严格配置读取未破坏联合路径。
4. 运行 miniobsidian.nvim 完整 CI，确认可选 Adapter 兼容。
5. 对照审计逐项记录：
   - 原问题；
   - 修复提交；
   - 测试证据；
   - 最终状态；
   - 保留风险。
6. 对未采纳项记录“不修改”理由，防止后续重复实施。
7. 更新 P5 子任务、阶段 README 和根计划状态。
8. 检查 P4 仍真实反映移动端门禁，不因 P5 完成而误标正式发布。

## 验收标准

- [ ] P5-T01 至 P5-T06 全部满足各自验收标准。
- [ ] `make release-check` 全部通过且覆盖率不低于当前门槛。
- [ ] 三入口自动 E2E 输出 `6/6`。
- [ ] miniobsidian.nvim CI 全部通过。
- [ ] 两个仓库没有未解释的工作区修改。
- [ ] 审计报告能从问题追踪到提交和测试。
- [ ] P4 移动端待办与 release-candidate 状态保持不变。

## 验证命令

```bash
cd /Users/andy/github/obs-cli
make release-check
./scripts/run-three-client-e2e.sh
git diff --check
git status --short

cd /Users/andy/github/miniobsidian.nvim
make ci PLENARY_DIR=/Users/andy/.local/share/nvim/lazy/plenary.nvim
git status --short
```

## 回滚

T07 只提交报告和状态。如联合回归失败，不回滚其他任务掩盖问题，而是保持 T07
未完成并为失败项建立独立修复任务。
