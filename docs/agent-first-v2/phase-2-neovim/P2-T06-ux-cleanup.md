# P2-T06：CWD、配置、Health 与文档收口

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`miniobsidian.nvim`
- 依赖：P2-T01 至 P2-T05

## 目标

清除隐式副作用、配置残留和文档矛盾，完成独立插件的稳定发布面。

## 实施步骤

1. 增加 `change_cwd_on_switch`，默认不修改全局 cwd。
2. 需要 cwd 时优先使用 tab-local `tcd` 或显式回调。
3. 修复自动发现有效但 `vaults_parent` 为空时的 health 误报。
4. 每次 `setup()` 从 defaults 重新构造配置，避免历史值残留。
5. 清除实际不存在的 obs-cli 默认配置注释。
6. 统一 `:ObsidianNew[!]`、`:ObsidianToday[!]` 的文档和实际行为。
7. 明确 search/switch 的 `notes_subdir` 与全 Vault 范围选项。
8. 更新 README、health 输出和迁移说明。

## 交付物

- CWD 与配置行为修复
- health 修复
- 更新后的用户文档

## 验收标准

- [ ] 切换 Vault 默认不影响其他项目的全局 cwd。
- [ ] 零配置自动发现时 health 不报错误。
- [ ] 重复 setup 得到可预测配置。
- [ ] 所有命令帮助、注释和行为一致。
- [ ] 无 CLI 环境下完整 smoke test 通过。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
stylua --check .
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests" -c qa
```
