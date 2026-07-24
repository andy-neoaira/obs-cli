# P4-T02：Neovim 安全移动与批量操作入口

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`miniobsidian.nvim`、`obs-cli`
- 依赖：P4-T01、P1-T07

## 目标

让 Neovim 用户可选使用 CLI 的事务化移动、链接重写和审计能力，而不在插件中复制复杂实现。

## 实施步骤

1. 增加 `:ObsidianMove`，仅在 capability 满足时注册或启用。
2. 将当前 buffer 路径和 revision 传给 CLI dry-run。
3. 用 picker/preview 展示目标和所有链接 diff。
4. 用户确认后执行 apply。
5. apply 前检查 buffer 是否有未保存修改。
6. 成功后处理当前 buffer rename、缓存失效和窗口状态。
7. 失败时展示稳定错误和恢复信息。
8. 为 Vault Audit/批量操作提供只读结果入口，不自动应用。

## 交付物

- 安全移动 UI
- dry-run diff preview
- buffer/窗口同步逻辑

## 验收标准

- [ ] 未保存 buffer 时禁止移动。
- [ ] 用户确认前 Vault 无变化。
- [ ] 目标已存在和 revision 冲突有清晰提示。
- [ ] 成功后 buffer 指向新路径。
- [ ] 无 CLI 时不影响插件原生功能。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/integration/move" -c qa
```

