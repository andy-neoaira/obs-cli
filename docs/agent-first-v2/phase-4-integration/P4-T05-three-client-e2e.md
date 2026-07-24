# P4-T05：三入口端到端测试

- 状态：`未开始`
- 负责人：`待分配`
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
- E2E fixture 和黄金摘要
- 人工 smoke test 清单与记录

## 验收标准

- [ ] 所有自动场景可重复运行且不依赖个人配置。
- [ ] 任何冲突都不会静默覆盖。
- [ ] Daily Note、链接和 Frontmatter 在三个入口一致。
- [ ] 移动后不存在新增坏链接。
- [ ] 人工 Obsidian 桌面端 smoke test 通过。
- [ ] 移动端同步限制和观察结果有记录。

## 验证

```bash
cd /Users/andy/github/obs-cli
./scripts/run-three-client-e2e.sh
```

