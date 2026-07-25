# P2：miniobsidian.nvim 可靠性改造

## 阶段目标

保持插件独立和原生体验，同时让它在路径、Wikilink、Daily Note、外部修改等方面遵守共同规范。

## 进入条件

- P0-T01、P0-T02、P0-T04、P0-T05 已完成。
- 本地仓库 `/Users/andy/github/miniobsidian.nvim` 可用。

## 任务进度

阶段进度：`6 / 6`

- [x] [P2-T01 测试框架、格式化与 CI](./P2-T01-tests-ci.md)
- [x] [P2-T02 Vault 路径安全模块](./P2-T02-safe-paths.md)
- [x] [P2-T03 Wikilink 解析、消歧与跳转](./P2-T03-wikilinks.md)
- [x] [P2-T04 Daily Note 与模板一致性](./P2-T04-daily-templates.md)
- [x] [P2-T05 外部修改检测与缓存失效](./P2-T05-external-changes.md)
- [x] [P2-T06 CWD、配置、Health 与文档收口](./P2-T06-ux-cleanup.md)

推荐顺序：P2-T01 → P2-T02 → P2-T03/P2-T04/P2-T05 → P2-T06。

## 阶段完成标准

- [x] 无 CLI 时所有基础功能完整可用。
- [x] 共享 fixture 在 Lua 测试中通过。
- [x] 不存在 Vault 路径逃逸和同名笔记静默误选。
- [x] Daily Note 与 Obsidian/CLI 语义一致。
- [x] 外部修改不会静默覆盖未保存 buffer。
- [x] Stylua、静态检查、headless 测试和 CI 通过。

## 阶段验证

```bash
cd /Users/andy/github/miniobsidian.nvim
stylua --check .
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests" -c qa
```
