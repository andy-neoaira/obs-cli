# P3-T01：Skill 契约、模板与安全基线

- 状态：`未开始`
- 负责人：`待分配`
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

- [ ] 模板能表达只读和修改型 Skill。
- [ ] 明确哪些操作需要用户确认。
- [ ] 明确冲突后不得无条件重试覆盖。
- [ ] 明确修改后的读取验证。
- [ ] 现有 Skill lint 能报告缺失字段。

## 验证

```bash
./scripts/lint-skills.sh
```

