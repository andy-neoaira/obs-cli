# P3-T07：新增 Project Status 与 Safe Note Update

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli Skills`
- 依赖：P3-T01、P3-T03、P3-T06

## 目标

提供项目状态汇总和通用安全更新工作流，作为 Agent 写回笔记的标准范式。

## 实施步骤

1. 创建 `obsidian-project-status`，读取项目笔记、Daily Notes 和任务。
2. 比较上次状态，生成进展、风险、决策和下一步。
3. 创建 `obsidian-safe-note-update`。
4. 固化 read + revision → analyze → dry-run patch → apply → verify。
5. 对 `REVISION_CONFLICT` 提供重新读取、三方比较和放弃三个分支。
6. 禁止默认 whole-file replace。
7. 输出更新前后 revision 和实际 diff 摘要。
8. 增加 Obsidian/Neovim 外部修改发生在 plan/apply 之间的测试。

## 交付物

- 两个新 Skill
- 标准安全更新流程
- 并发冲突场景测试

## 验收标准

- [x] Project Status 不重复生成已有周期章节。
- [x] Safe Update 对冲突默认停止。
- [x] 未经显式授权不使用强制覆盖。
- [x] 修改范围与 dry-run 一致。
- [x] verify 能检测预期修改未生效。

## 验证

```bash
./scripts/lint-skills.sh
go test ./... -run 'Skill.*(ProjectStatus|SafeUpdate|Conflict)'
GOCACHE=/private/tmp/obs-cli-go-cache make release-check
```

## 完成记录

- 新增 `obsidian-project-status`，以 ISO week、显式时区和有界来源集生成
  progress/risks/decisions/next_steps，并保留 source → evidence → item 证据链。
- 写回只允许向唯一 `## Status` 追加唯一周期块；同周期完全一致返回
  `no_change`，内容不同或残缺返回 `conflict`，不生成重复章节。
- 新增 `obsidian-safe-note-update`，固定 read → plan → dry-run → authorize →
  apply → verify，禁止 whole-file replace、frontmatter patch 和无 revision 写入。
- apply 使用受控 match/replacement 文件 digest、规范化 plan digest 和稳定 request
  ID；verify 比较完整 expected bytes、planned/apply revision 及无关字节。
- 冲突默认停止，并定义 reread/three-way/abandon、未知结果查询与单次幂等重放。
- 新增 Draft 2020-12 Project Status Schema，以及只读、applied、no_change、
  conflict 四份黄金输出；全量验证来源、证据、baseline、条目与写回语义。
- 场景测试覆盖 Status 章节边界、周期去重/冲突、dry-run/apply/verify revision
  一致、外部修改冲突和 plan 后受控输入被替换。
- 两轮独立前向测试通过；完整 release-check 通过，覆盖率 72.3%。
