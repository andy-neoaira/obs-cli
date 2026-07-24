# P3-T03：迁移 Daily Log 与 Project Note

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`obs-cli Skills`
- 依赖：P3-T01、P1-T06、P2-T04

## 目标

迁移每日记录和项目笔记工作流，并遵守共同 Daily/Frontmatter 规范。

## 实施步骤

1. Daily Log 使用 `daily get/create/append`。
2. 创建前解析官方 Daily 配置和模板。
3. 追加到章节时使用 section-aware patch。
4. Project Note 读取当前 revision 和 Frontmatter。
5. 仅更新目标 metadata 字段或章节，避免整篇覆盖。
6. dry-run 展示新增章节、字段和正文差异。
7. 冲突时重新读取并重新规划，不自动强制写入。
8. 增加跨入口创建相同 Daily Note 的测试。

## 交付物

- 更新后的两个 `SKILL.md`
- Daily/Project 场景测试

## 验收标准

- [ ] Daily Note 路径与插件/Obsidian 一致。
- [ ] 重复执行不会重复创建同一章节。
- [ ] Project metadata 不破坏未知 Frontmatter 字段。
- [ ] revision 冲突可被正确报告。
- [ ] 修改后验证目标章节和字段。

## 验证

```bash
./scripts/lint-skills.sh
go test ./... -run 'Skill.*(Daily|Project)'
```

