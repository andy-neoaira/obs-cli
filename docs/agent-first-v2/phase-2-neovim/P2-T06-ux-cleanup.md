# P2-T06：CWD、配置、Health 与文档收口

- 状态：`已完成`
- 负责人：`Codex`
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

- [x] 切换 Vault 默认不影响其他项目的全局 cwd。
- [x] 零配置自动发现时 health 不报错误。
- [x] 重复 setup 得到可预测配置。
- [x] 所有命令帮助、注释和行为一致。
- [x] 无 CLI 环境下完整 smoke test 通过。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
PATH=/private/tmp/miniobsidian-tools/bin:$PATH \
  make ci PLENARY_DIR=/Users/andy/.local/share/nvim/lazy/plenary.nvim
make test \
  NVIM=/private/tmp/nvim-macos-arm64/bin/nvim \
  PLENARY_DIR=/Users/andy/.local/share/nvim/lazy/plenary.nvim
```

## 完成记录

- Vault 切换默认不再修改 cwd；显式启用后仅使用 tab-local `tcd`。
- `setup()` 每次从默认值重建配置，并统一校验枚举、时间、格式和 Vault 相对路径。
- health 支持零配置自动发现，缺少可选 picker/`rg` 时降级为可操作警告。
- picker 范围可在 `notes_subdir` 与整个 Vault 间显式选择。
- 中英文 README、命令帮助、最低 Neovim 版本和 V2 迁移说明已统一。
- 默认 Neovim 与 0.10.4 的全量 headless 测试均通过；Stylua、Selene、fixture 校验通过。
