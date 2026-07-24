# P4-T01：miniobsidian 可选 CLI Adapter

- 状态：`未开始`
- 负责人：`待分配`
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

- [ ] 默认安装不要求 obs-cli。
- [ ] CLI 调用不会阻塞 Neovim UI。
- [ ] 参数不经过 shell 插值。
- [ ] 协议不兼容不会调用修改操作。
- [ ] 基础笔记功能在 CLI 失败后继续可用。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/cli" -c qa
```

