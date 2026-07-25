# P3：场景化 Agent Skills

## 阶段目标

将 Skill 从“命令用法说明”升级为具备前置检查、计划、写入授权、冲突处理和结果验证的场景工作流。

## 进入条件

- P1 已完成并提供稳定 `obs-cli/v2`。
- `capabilities`、JSON、dry-run、revision 和错误码可用。

## 任务进度

阶段进度：`3 / 8`

- [x] [P3-T01 Skill 契约、模板与安全基线](./P3-T01-skill-contract.md)
- [x] [P3-T02 迁移 Vault Setup 与 Capture](./P3-T02-setup-capture.md)
- [x] [P3-T03 迁移 Daily Log 与 Project Note](./P3-T03-daily-project.md)
- [ ] [P3-T04 迁移 Knowledge Search 与 Vault Audit](./P3-T04-search-audit.md)
- [ ] [P3-T05 迁移 Inbox Triage](./P3-T05-inbox-triage.md)
- [ ] [P3-T06 新增 Compare 与 Knowledge Synthesis](./P3-T06-compare-synthesis.md)
- [ ] [P3-T07 新增 Project Status 与 Safe Note Update](./P3-T07-project-safe-update.md)
- [ ] [P3-T08 Skill 场景评测与发布门禁](./P3-T08-skill-evals.md)

推荐顺序：P3-T01 → P3-T02/P3-T03/P3-T04/P3-T05 → P3-T06 → P3-T07 → P3-T08。

## 阶段完成标准

- [ ] 所有 Skill 执行前检查 capability。
- [ ] 所有修改型 Skill 先 dry-run，再按授权策略执行。
- [ ] 所有更新型 Skill处理 `REVISION_CONFLICT`。
- [ ] 每个 Skill 有触发、非触发、成功、冲突和失败测试。
- [ ] Skill 输出包含修改摘要和验证结果。

## 阶段验证

```bash
cd /Users/andy/github/obs-cli
find skills -name SKILL.md -print
go test ./... -run 'Skill|Scenario|Eval'
```
