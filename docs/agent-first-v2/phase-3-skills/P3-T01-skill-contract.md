# P3-T01：Skill 契约、模板与安全基线

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli Skills`
- 依赖：P1-T05、P1-T06

## 目标

定义所有场景 Skill 必须遵守的统一结构和执行生命周期。

## 实施步骤

1. 创建 `skills/_template/SKILL.md`。
2. 必填章节包括触发、非触发、输入、capability、读取范围、写入范围。
3. 定义 `discover → read → plan → authorize → apply → verify` 生命周期。
4. 定义 dry-run、revision、幂等、重试和部分失败规则。
5. 定义 shell 安全规则：长文本只通过 stdin 或文件传入。
6. 定义结构化结果摘要格式。
7. 创建 Skill lint 脚本，检查必填章节和 capability 声明。

## 交付物

- `skills/_template/SKILL.md`
- `docs/spec/SKILL_CONTRACT.md`
- Skill lint 工具

## 验收标准

- [x] 模板能表达只读和修改型 Skill。
- [x] 明确哪些操作需要用户确认。
- [x] 明确冲突后不得无条件重试覆盖。
- [x] 明确修改后的读取验证。
- [x] 现有 Skill lint 能报告缺失字段。

## 验证

```bash
./scripts/lint-skills.sh
./scripts/lint-skills.sh skills/_template/SKILL.md
python3 /Users/andy/.codex/skills/.system/skill-creator/scripts/quick_validate.py \
  skills/_template
GOCACHE=/private/tmp/easyskill-go-cache make release-check
```

## 完成记录

- 新增 V2 Skill 模板，统一触发/非触发、输入、capability、读写范围和结果摘要。
- 固化 `discover → read → plan → authorize → apply → verify` 生命周期。
- 明确 dry-run、revision、冲突停止、幂等重试、部分失败和 shell 长文本安全规则。
- 新增 lint 工具与 Makefile 发布门禁；默认渐进报告 V1 Skill，指定文件或 `--strict` 时阻断不合规项。
- 使用缺失“写入范围”的坏样本验证 lint 返回非零并准确报告字段。
- Skill 模板校验、完整 release-check 与 RC smoke 均通过。
