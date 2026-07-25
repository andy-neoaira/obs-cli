# P3-T04：迁移 Knowledge Search 与 Vault Audit

- 状态：`已完成`
- 负责人：`Codex`
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

- [x] 两个 Skill 默认不写 Vault。
- [x] 结论可以追溯到具体笔记和 revision。
- [x] 搜索读取量有上限。
- [x] 审计不会把附件当 Markdown 解析。
- [x] 无结果与命令失败能够区分。

## 验证

```bash
./scripts/lint-skills.sh
go test ./... -run 'Skill.*(Search|Audit)'
GOCACHE=/private/tmp/easyskill-go-cache make release-check
```

## 完成记录

- 新增 `search.content`：scope、分页、page-size/max-files 硬上限及 path/revision/size/line/snippet 证据。
- 新增 `link.backlinks`：支持相同 scope 边界，返回来源 path/revision/line 和链接类型。
- raw search 不解析 frontmatter；坏 YAML 仍可搜索，`note get` 的错误 details 提供 path/revision。
- 非 Markdown 附件不会进入扫描；附件链接无法枚举时只报告 unverified，不误报 broken link。
- 新增 search/audit 报告 Schema，区分 query、positive evidence、negative checks、per-type coverage、两类截断、错误和未执行修复 plan。
- Knowledge Search 与 Vault Audit 已迁移为有界只读 V2 工作流。
- 空结果、大 Vault 截断、非法 page-size、附件排除、坏 frontmatter、backlinks scope/revision 场景测试通过。
- 两轮独立前向测试完成；由于缺少 Vault-wide snapshot，orphan 安全降级为 candidate。
