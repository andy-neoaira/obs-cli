# P4-T02：Neovim 安全移动与批量操作入口

- 状态：`已完成`
- 负责人：`Codex`
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

- [x] 未保存 buffer 时禁止移动。
- [x] 用户确认前 Vault 无变化。
- [x] 目标已存在和 revision 冲突有清晰提示。
- [x] 成功后 buffer 指向新路径。
- [x] 无 CLI 时不影响插件原生功能。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/integration/move" -c qa
PATH=/private/tmp/miniobsidian-tools/bin:$PATH make ci \
  PLENARY_DIR=/Users/andy/.local/share/nvim/lazy/plenary.nvim
make test NVIM=/private/tmp/nvim-macos-arm64/bin/nvim \
  PLENARY_DIR=/Users/andy/.local/share/nvim/lazy/plenary.nvim
```

## 完成记录

- `miniobsidian.nvim` 提交：`72781f0 feat: add safe cli move workflow`。
- 新增 `miniobsidian.move` 与 `:ObsidianMove [target]`：只在 `note.get`、
  `note.move` capability 就绪时启动，先读取 revision，再展示 dry-run 计划。
- 用户选择 `Apply` 后才携带相同 `revision + plan_hash` 执行 mutation；取消确认、
  未保存 buffer、确认期间 buffer 变化都不会启动 apply。
- 对 apply 回执校验 `plan_hash` 和目标路径；成功后同步原 buffer 到新路径、
  强制重载磁盘内容并失效笔记缓存。
- 稳定展示 `ALREADY_EXISTS`、`REVISION_CONFLICT`、capability/协议/JSON 等错误；
  CLI 不可用时原生笔记功能保持独立。
- 新增 `:ObsidianVaultAudit`：通过只读 `note.list` 打开不可修改的 JSON 快照，
  不自动执行审计或批量写入。
- fake CLI 测试覆盖未保存阻断、dry-run 取消、目标已存在、成功移动与 buffer
  同步、revision 冲突、只读审计共 6 个场景。
- 隔离副本和实际插件仓库的完整 CI 均通过；Neovim 0.10.4 全量测试通过。
