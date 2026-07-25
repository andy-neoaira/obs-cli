# P4-T03：当前笔记与选区交给 Agent

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`miniobsidian.nvim`、Agent 集成
- 依赖：P4-T01、P3-T06、P3-T07

## 目标

允许用户从 Neovim 将当前笔记、选区和任务意图安全地交给 Agent 分析或更新。

## 实施步骤

1. 定义 handoff payload：Vault ID、相对路径、revision、选区、意图。
2. 默认仅传路径和选区，不隐式传整个 Vault。
3. 增加 `:ObsidianAgentAnalyze` 和 `:ObsidianAgentUpdate`。
4. buffer 有未保存修改时，要求保存或仅发送内存内容做只读分析。
5. 更新型 handoff 必须指定允许写入的目标路径。
6. Agent 使用对应 Skill 和 capability 检查。
7. 保存 request ID，便于关联执行结果。
8. 对敏感内容和大选区增加预览/确认配置。

## 交付物

- handoff payload 规范
- Neovim 命令与 Agent 接口
- 权限边界测试

## 验收标准

- [x] Agent 获得明确路径和 revision。
- [x] 默认不会读取未授权的整个 Vault。
- [x] 未保存 buffer 的行为明确且无数据丢失。
- [x] 更新请求能追踪到 request ID。
- [x] 只读分析不会触发文件修改。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/integration/handoff" -c qa
PATH=/private/tmp/miniobsidian-tools/bin:$PATH make ci \
  PLENARY_DIR=/Users/andy/.local/share/nvim/lazy/plenary.nvim
make test NVIM=/private/tmp/nvim-macos-arm64/bin/nvim \
  PLENARY_DIR=/Users/andy/.local/share/nvim/lazy/plenary.nvim
```

## 完成记录

- `miniobsidian.nvim` 提交：`c88952f feat: add bounded agent handoff`。
- 定义 `miniobsidian.agent-handoff/v1` JSON Schema 和边界规范：只传稳定 Vault ID、
  当前笔记相对路径/revision、显式意图及可选选区，不传本机绝对路径，默认
  `allow_vault_scan=false`。
- 新增 `agent.handler(payload)` 适配接口，不绑定具体 Agent/聊天框架，不执行 shell；
  handler 成功后保存 `last_request` 并触发仅含 request ID/mode/path 的 User 事件。
- 新增 `:ObsidianAgentAnalyze` 与 `:ObsidianAgentUpdate`，均支持显式行范围和意图；
  未提供意图时通过 `vim.ui.input` 获取。
- analyze 绑定 `obsidian-knowledge-synthesis` 与只读路径；update 绑定
  `obsidian-safe-note-update`、`note.get + note.patch` capability，并只授权当前路径。
- dirty buffer 禁止 update；analyze 可发送选区或完整内存快照，但内容发送前默认
  预览确认，大选区即使关闭普通确认仍强制确认。
- handoff 前通过 `note.get` 获取 Vault ID 与 revision，校验返回 path；更新请求在
  异步快照返回后再次检查 buffer，避免构建期间变脏。
- Schema 测试验证 analyze/update 正例，并拒绝只读写权限、dirty update 和全
  Vault 扫描；fake CLI 专项测试覆盖权限、取消、dirty buffer、Skill/capability、
  handler 缺失和大选区共 7 个场景。
- 实际插件仓库完整 CI、Neovim 0.10.4 全量测试及最终专项测试均通过。
