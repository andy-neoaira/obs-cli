# P3-T03：迁移 Daily Log 与 Project Note

- 状态：`已完成`
- 负责人：`Codex`
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

- [x] Daily Note 路径与插件/Obsidian 一致。
- [x] 重复执行不会重复创建同一章节。
- [x] Project metadata 不破坏未知 Frontmatter 字段。
- [x] revision 冲突可被正确报告。
- [x] 修改后验证目标章节和字段。

## 验证

```bash
./scripts/lint-skills.sh
go test ./... -run 'Skill.*(Daily|Project)'
GOCACHE=/private/tmp/easyskill-go-cache make release-check
```

## 完成记录

- 补齐任务依赖的 `daily.get/create/append` 与 `metadata.get/set` V2 operation。
- Daily 读取官方 folder/format/template，create 不覆盖其他入口已创建的日记，append 强制 revision。
- Metadata 单字段更新复用原子写入，只改变目标键并语义保留其他字段和正文。
- capabilities、feature flags、命令树与协议文档已同步。
- Daily/Project Skills 已迁移到分支化 capability、逐步 dry-run、冲突停止、partial 恢复和写后 delta 验证。
- 增加模板渲染、嵌套日期路径、section append、stale revision、未知 metadata 保留及跨入口创建测试。
- 两个 Skill 通过 lint、结构校验和独立前向测试。
