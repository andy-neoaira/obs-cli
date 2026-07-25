# P4-T01：miniobsidian 可选 CLI Adapter

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`miniobsidian.nvim`、`obs-cli`
- 依赖：P1、P2

## 目标

为插件提供可选 CLI 能力探测和调用层，同时确保 CLI 缺失、版本不符或执行失败时基础功能不受影响。

## 实施步骤

1. 增加 `cli.enabled = false|auto|true` 和 `cli.command` 配置。
2. `auto` 模式异步检查 executable 和 `capabilities --output json`。
3. 缓存 capability，提供显式刷新。
4. 封装 argv 调用，禁止拼接 shell 命令字符串。
5. 解析 JSON envelope 和稳定错误码。
6. CLI 不可用时隐藏高级命令并在 health 中给出可选提示。
7. CLI 超时、非法 JSON 和协议不兼容时安全降级。
8. 增加 fake CLI 测试。

## 交付物

- `lua/miniobsidian/cli.lua`
- CLI 配置和 health 项
- fake CLI 集成测试

## 验收标准

- [x] 默认安装不要求 obs-cli。
- [x] CLI 调用不会阻塞 Neovim UI。
- [x] 参数不经过 shell 插值。
- [x] 协议不兼容不会调用修改操作。
- [x] 基础笔记功能在 CLI 失败后继续可用。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/cli" -c qa
PATH=/private/tmp/miniobsidian-tools/bin:$PATH make ci \
  PLENARY_DIR=/Users/andy/.local/share/nvim/lazy/plenary.nvim
make test NVIM=/private/tmp/nvim-macos-arm64/bin/nvim \
  PLENARY_DIR=/Users/andy/.local/share/nvim/lazy/plenary.nvim
```

## 完成记录

- `miniobsidian.nvim` 提交：`4e56ddf feat: add optional obs-cli adapter`。
- 新增 `cli.enabled = false | "auto" | true`、`cli.command` 和 `cli.timeout_ms`；
  默认 `auto`，CLI 缺失时插件本地功能不降级。
- 新增 `miniobsidian.cli`：使用 `vim.system(argv)` 异步探测
  `capabilities --output json`，缓存 capability，并提供显式 refresh/call/state/
  available API。
- 只接受 `obs-cli/v2` envelope 和 capability 声明；超时、进程失败、非法 JSON、
  协议不兼容和稳定 CLI error 都转换为结构化 adapter 状态。
- CLI 调用只接受字符串 argv table，不拼接 shell；协议不兼容时 mutation 在进程
  启动前被拒绝。
- 新增 `:ObsidianCLIRefresh` 与 `:checkhealth miniobsidian` 可选状态；`auto`
  缺失仅为 info，显式 `true` 失败为 warn。
- fake CLI 测试覆盖缺失 executable、缓存/刷新、argv shell 注入、协议不兼容、
  timeout、非法 JSON，以及失败后本地扫描仍可用。
- 实际插件仓库完整 CI 与 Neovim 0.10.4 全量测试均通过。
