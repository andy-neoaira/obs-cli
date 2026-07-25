# P2-T04：Daily Note 与模板一致性

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`miniobsidian.nvim`、`obs-cli`（共同契约补齐）
- 依赖：P0-T02、P0-T05、P2-T01、P2-T02

## 目标

让插件创建的 Daily Note 与 Obsidian 官方 Daily Notes 配置及共同规范一致。

## 实施步骤

1. 扩展配置同步，读取 Daily Notes 的 folder、format 和 template。
2. 已配置模板时读取模板并执行规范允许的变量替换。
3. 未配置模板时使用可配置的最小默认内容。
4. 使用日历日期计算 yesterday/tomorrow，移除固定 `±86400`。
5. 模板选择使用相对路径，禁止 basename 静默覆盖。
6. 已存在 Daily Note 只打开，不覆盖。
7. 用共同 fixture 比较 CLI 与插件生成的目标路径和内容。

## 交付物

- Daily 配置同步
- 模板解析修复
- 语义一致性测试

## 验收标准

- [x] Obsidian、CLI、插件计算出相同 Daily Note 路径。
- [x] 配置模板被正确使用。
- [x] 同名模板可消歧。
- [x] 夏令时边界日期计算正确。
- [x] 已存在文件内容不被覆盖。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/daily tests/template" -c qa
```

## 完成记录

- 2026-07-25：插件同步 `.obsidian/daily-notes.json` 的 `folder`、`format`、`template`，默认目录改为 Vault 根目录，无模板时默认创建空文件。
- 2026-07-25：新增共同模板渲染器，变量大小写不敏感，未知变量保留并 warning，日期格式遇到不支持 token 时明确失败。
- 2026-07-25：`yesterday`/`tomorrow` 改为本地日历日期运算，并用 America/New_York 夏令时切换测试证明不依赖固定 `±86400`。
- 2026-07-25：模板选择器使用模板目录相对路径；Daily 配置模板按唯一 Note ID 解析，缺失、歧义和路径逃逸均停止创建。
- 2026-07-25：已存在 Daily Note 直接打开且不读取已移除的历史模板，不修改原有字节。
- 2026-07-25：共同 fixture `vault-contract/v1/daily-template` 同时覆盖 Go 与 Lua，实现路径、模板内容与未知变量语义一致。
- 2026-07-25：发现 CLI 旧 Daily 实现未渲染模板且会忽略模板读取错误；作为共同契约前置缺口在本任务中一并修复。
- 2026-07-25：`miniobsidian.nvim` 完整 CI、Neovim 0.10.4 回归及 `obs-cli make release-check` 全部通过。
