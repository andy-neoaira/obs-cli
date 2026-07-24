# P3-T04：迁移 Knowledge Search 与 Vault Audit

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`obs-cli Skills`
- 依赖：P3-T01、P1-T08

## 目标

将搜索与审计明确为只读场景，提供可追溯的证据和分页策略。

## 实施步骤

1. Knowledge Search 使用 V2 search/note get/link backlinks。
2. 定义查询扩展、分页、最大读取量和停止条件。
3. 输出引用笔记路径、revision 和匹配片段。
4. Vault Audit 检查坏链接、孤立笔记、重复标题、Frontmatter 和 TODO。
5. 审计只生成报告，不自动修复。
6. 修复建议输出为可供后续 Skill 使用的 plan。
7. 增加无结果、大 Vault、二进制附件和坏 Frontmatter 场景。

## 交付物

- 更新后的两个 `SKILL.md`
- 搜索与审计报告 Schema
- 场景测试

## 验收标准

- [ ] 两个 Skill 默认不写 Vault。
- [ ] 结论可以追溯到具体笔记和 revision。
- [ ] 搜索读取量有上限。
- [ ] 审计不会把附件当 Markdown 解析。
- [ ] 无结果与命令失败能够区分。

## 验证

```bash
./scripts/lint-skills.sh
go test ./... -run 'Skill.*(Search|Audit)'
```

