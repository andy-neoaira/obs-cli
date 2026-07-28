# P5-T06：Skill 身份三方一致性门禁

- 状态：`已完成`
- 优先级：`中`
- 负责人：`待分配`
- 涉及项目：`obs-cli`、Agent Skills
- 依赖：P3-T01、P3-T08
- 来源：审计 m7

## 背景与已确认现象

Skill lint 校验 frontmatter `name` 格式，但没有验证它与目录名一致。
eval 测试通过 manifest name 定位目录，也没有比较文件中的 frontmatter name。

## 目标

保证正式 Skill 的目录名、frontmatter `name` 和
`skills/evals/scenarios.json` 中的 name 完全一致。

## 非目标

- 不修改现有 Skill 的触发语义和场景数量。
- 不要求模板目录 `_template` 的目录名等于示例 name。
- 不引入完整 YAML 运行时依赖。

## 修改范围

- `scripts/lint-skills.sh`
- `cmd/skill_evals_test.go`
- `skills/evals/scenarios.json` 和 Schema（仅在需要补充约束时）
- Skill lint 负向 fixture

## 执行过程

1. 在 lint 中可靠提取 frontmatter 的唯一 `name`。
2. 对 `skills/<name>/SKILL.md` 比较目录名与 frontmatter name。
3. 对 `_template` 使用明确豁免，同时继续校验模板结构。
4. 在 eval 测试中比较 manifest name、目录名和 frontmatter name。
5. 增加以下负向用例：
   - 合法格式但与目录不一致；
   - manifest 中存在但目录缺失；
   - 目录存在但 manifest 遗漏；
   - 重复 name。
6. 保证错误信息包含 Skill 路径和三方实际值。

## 验收标准

- [x] 任意正式 Skill 名称不一致都会使 `skill-lint` 失败。
- [x] manifest 缺项、多项或重复项都会使 eval 门禁失败。
- [x] `_template` 不被错误要求命名为 `_template`。
- [x] 当前所有正式 Skill 和 5 类场景 eval 继续通过。
- [x] lint 不依赖文件遍历的非确定顺序。

## 验证命令

```bash
./scripts/lint-skills.sh --strict
./scripts/run-skill-evals.sh
go test ./cmd -run 'Skill.*Eval|Skill.*Identity'
make release-check
git diff --check
```

## 回滚

整体回滚 lint、eval 测试和 fixture；回滚后必须运行现有 Skill lint/eval，
确认没有留下只由新门禁支持的 manifest 结构。

## 完成记录

- 完成日期：`2026-07-28`
- lint：正式 Skill 的 frontmatter name 必须等于目录名，`_template` 明确豁免。
- eval：新增 manifest name 唯一性和 frontmatter/目录/manifest 三方比较；既有
  `skillDirectories` 覆盖 manifest 缺项和多项。
- 验证：strict Skill lint、场景 eval 和 `make release-check` 通过；总覆盖率
  `73.8%`。
