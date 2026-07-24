# P2-T01：测试框架、格式化与 CI

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`miniobsidian.nvim`
- 依赖：P0-T05

## 目标

为插件建立可重复运行的 headless 测试和持续集成，覆盖所有写文件模块。

## 实施步骤

1. 选择并固定 Lua/Neovim 测试框架，建立 `tests/minimal_init.lua`。
2. 为 `init`、`vault`、`note`、`link`、`daily`、`template` 建立基础测试。
3. 使用临时 Vault，禁止测试读写真实个人 Vault。
4. 配置 Stylua 和 Selene 检查。
5. 修复当前全部 Stylua 差异。
6. 新建 GitHub Actions，覆盖最低支持版本和当前稳定版 Neovim。
7. 增加插件 `require` 与 `setup` smoke test。
8. 在 CI 中校验共享 fixture 的版本、内容摘要和同步状态。

## 交付物

- `tests/`
- `.github/workflows/ci.yml`
- 通过格式化的 Lua 源码

## 验收标准

- [ ] 测试可在干净环境运行。
- [ ] 测试不会访问用户 Obsidian 配置和真实 Vault。
- [ ] Stylua 检查无差异。
- [ ] 最低支持版和稳定版 Neovim CI 通过。
- [ ] CI 能在共享 fixture 版本或内容未同步时失败。
- [ ] 测试失败时 CI 返回非零状态。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
stylua --check .
selene lua
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests" -c qa
```
