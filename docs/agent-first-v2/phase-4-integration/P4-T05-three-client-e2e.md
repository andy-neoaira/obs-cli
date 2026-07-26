# P4-T05：三入口端到端测试

- 状态：`进行中（自动化 6/6 与桌面端已通过，待移动端同步观察）`
- 负责人：`Codex（自动化与桌面端）/ 待分配（移动端）`
- 涉及项目：`obs-cli`、`miniobsidian.nvim`、Obsidian 兼容语义
- 依赖：P4-T01 至 P4-T04

## 目标

用自动化文件事件模拟和人工 Obsidian smoke test 验证完整信息闭环。

## 实施步骤

1. 创建隔离的 E2E Vault，禁止使用个人 Vault。
2. 场景 A：Obsidian 风格创建 → Agent 搜索/整理 → Neovim 读取。
3. 场景 B：Neovim 创建项目笔记 → Agent 生成状态更新。
4. 场景 C：Agent plan 后模拟外部修改 → apply 返回冲突。
5. 场景 D：CLI 移动并重写链接 → 插件刷新 buffer/cache。
6. 场景 E：三个入口创建/打开同一 Daily Note。
7. 场景 F：dirty buffer 与 Agent 外部更新三方比较。
8. 自动测试文件层行为，人工 smoke test 记录真实桌面端/移动同步结果。

## 交付物

- `scripts/run-three-client-e2e.sh`
- `testdata/three-client/` E2E fixture 和黄金摘要
- `miniobsidian.nvim/tests/three_client_e2e_spec.lua`
- [人工 smoke test 清单与记录](./P4-T05-manual-smoke.md)

## 自动化实施记录

- CLI 使用 `OBS_CLI_CONFIG_HOME` 指向临时配置根，Vault registry 不读取个人配置。
- 每次执行复制合成 fixture 到 `mktemp` Vault，结束后清理。
- 使用当前源码构建真实 `obs-cli`，并通过真实 `miniobsidian.nvim` Adapter
  运行搜索、条件更新、冲突、移动、Daily Note 和 dirty buffer 三方比较。
- `--vault` 支持 ID、名称或已注册绝对路径；绝对路径不会注册或信任任意目录。
- 输出只比较稳定场景与不变量，排除 Vault ID、临时路径、时间戳和 revision。

## 验收标准

- [x] 所有自动场景可重复运行且不依赖个人配置。
- [x] 任何冲突都不会静默覆盖。
- [x] Daily Note、链接和 Frontmatter 在三个入口一致。
- [x] 移动后不存在新增坏链接。
- [x] 人工 Obsidian 桌面端 smoke test 通过。
- [ ] 移动端同步限制和观察结果有记录。

## 验证

```bash
cd /Users/andy/github/obs-cli
./scripts/run-three-client-e2e.sh
```

2026-07-25 自动验证结果：

- `./scripts/run-three-client-e2e.sh`：`6/6` 通过，黄金摘要一致。
- `GOCACHE=/private/tmp/obs-cli-p4-go-cache make release-check`：通过。
- `miniobsidian.nvim make ci`：format、Selene、fixture 与完整 Plenary 回归通过。
- Obsidian Desktop `1.12.7`：真实 UI 创建/刷新、CLI revision 写入与移动、
  Neovim 读取/创建、Daily Note 和 stale revision 冲突保护均通过；完整证据见
  [人工 smoke test 清单](./P4-T05-manual-smoke.md)。

移动端同步结果尚未填写，因此本任务和 P4 阶段仍不能标记完成。
