# P0-T01：产品边界与架构决策

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli`、`miniobsidian.nvim`
- 依赖：无

## 目标

用架构决策记录冻结三个入口的职责、依赖方向和非目标，防止后续重新引入 Human-first CLI 或让 Neovim 插件硬依赖 CLI。

## 实施步骤

1. 新建 `docs/architecture/ADR-001-agent-first-boundary.md`。
2. 记录 Vault 是唯一内容事实源。
3. 定义 Obsidian、`obs-cli`、Skills、`miniobsidian.nvim` 的职责。
4. 明确插件不强制依赖 CLI，CLI 也不依赖 Neovim。
5. 明确 Obsidian 官方配置只读发现、各工具自有配置独立存储。
6. 列出 V2 非目标：TTY 交互、内置 picker、自动启动编辑器、隐式打开应用。
7. 在两个项目 README 中链接 ADR。

## 交付物

- `obs-cli/docs/architecture/ADR-001-agent-first-boundary.md`
- 两个项目 README 的架构说明与链接

## 验收标准

- [x] ADR 包含背景、决策、替代方案、后果和状态。
- [x] 明确说明三入口是同级客户端。
- [x] 明确说明 `miniobsidian.nvim` 无 CLI 硬依赖。
- [x] 明确列出 Agent-first CLI 的非目标。
- [x] 两个仓库文档表述一致。

## 验证

```bash
rg -n "Agent-first|唯一内容事实源|不.*依赖" \
  /Users/andy/github/obs-cli/docs/architecture \
  /Users/andy/github/miniobsidian.nvim/README.md
```

## 验证记录

- 2026-07-24：关键词与 ADR 章节检查通过。
- 2026-07-24：`obs-cli` 本地 Markdown 链接检查通过，坏链接数为 0。
- 2026-07-24：两个仓库执行 `git diff --check`，均通过。
