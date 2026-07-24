# P3-T07：新增 Project Status 与 Safe Note Update

- 状态：`未开始`
- 负责人：`待分配`
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

- [ ] Project Status 不重复生成已有周期章节。
- [ ] Safe Update 对冲突默认停止。
- [ ] 未经显式授权不使用强制覆盖。
- [ ] 修改范围与 dry-run 一致。
- [ ] verify 能检测预期修改未生效。

## 验证

```bash
./scripts/lint-skills.sh
go test ./... -run 'Skill.*(ProjectStatus|SafeUpdate|Conflict)'
```

