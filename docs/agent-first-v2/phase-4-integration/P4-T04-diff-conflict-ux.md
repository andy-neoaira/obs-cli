# P4-T04：Agent 修改后的 Diff 与冲突 UX

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`miniobsidian.nvim`、`obs-cli`
- 依赖：P4-T03、P2-T05

## 目标

让用户在 Neovim 中清晰查看 Agent 的实际修改，并安全处理外部更新冲突。

## 实施步骤

1. Agent 响应返回 revision before/after、changed files 和摘要。
2. 插件收到结果后检查当前 buffer modified 状态。
3. 未修改 buffer 可执行 `checktime` 并展示 diff。
4. 已修改 buffer 打开三方比较：原 revision、Agent 结果、本地 buffer。
5. 提供保留本地、采用磁盘、手动合并动作。
6. 禁止自动选择冲突解决方案。
7. 多文件结果提供 changed-files picker。
8. 记录失败、取消和部分失败的用户可操作说明。

## 交付物

- Agent result handler
- Diff/三方比较 UI
- 冲突场景测试

## 验收标准

- [x] 用户能看到 Agent 实际修改了哪些文件。
- [x] dirty buffer 永不自动重载。
- [x] 冲突处理动作不会丢失任一版本。
- [x] `PARTIAL_FAILURE` 能展示恢复清单。
- [x] 取消操作不继续写入。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/integration/conflict" -c qa
PATH=/private/tmp/miniobsidian-tools/bin:$PATH make ci \
  PLENARY_DIR=/Users/andy/.local/share/nvim/lazy/plenary.nvim
make test NVIM=/private/tmp/nvim-macos-arm64/bin/nvim \
  PLENARY_DIR=/Users/andy/.local/share/nvim/lazy/plenary.nvim
```

## 完成记录

- `miniobsidian.nvim` 提交：`043f24a feat: add agent result conflict ux`。
- 定义 `miniobsidian.agent-result/v1` Schema：关联 handoff request ID，记录状态、
  changed files、revision before/after、摘要、可选 before content 和恢复步骤。
- 新增 `miniobsidian.agent_result.handle(result)`：拒绝非法或 request ID 不匹配的
  result，保存最近结果，并生成 changed files/revision/错误恢复摘要。
- `cancelled` 立即停止且不打开后续 UI；`partial/failed/conflict` 必须包含 errors，
  `PARTIAL_FAILURE` 等错误展示逐项目 recovery checklist。
- 多文件 result 通过 picker 选择检查目标；所有路径再次执行 Vault path policy，
  不信任 Agent 返回的相对路径。
- clean buffer 展示内存旧版本到 Agent 磁盘版本的 unified diff，并标记外部冲突、
  阻止陈旧写入；不会自动 reload。
- dirty buffer 打开 base / Agent disk / Neovim local 三方只读视图，并提供保留、
  采用磁盘、手动合并动作；不自动选择。采用磁盘前已将 local 保存为隐藏 scratch，
  手动合并只创建可编辑 scratch，不直接写 Vault。
- dirty result 缺少 `before_content` 时返回 `AGENT_BASE_MISSING`，保留 local/disk
  两个版本并继续阻止写入，不伪造 base。
- 新增 `:ObsidianAgentLastResult`；专项测试覆盖非法/错配、partial recovery、
  cancelled、多文件 picker、clean diff、dirty 三方和缺失 base 共 7 个场景。
- Schema 正例和 cancelled/partial/非 Markdown 负例通过；实际插件完整 CI 与
  Neovim 0.10.4 全量测试均通过。
