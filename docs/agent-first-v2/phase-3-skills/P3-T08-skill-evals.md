# P3-T08：Skill 场景评测与发布门禁

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`obs-cli Skills`
- 依赖：P3-T01 至 P3-T07

## 目标

建立可重复的 Skill 评测，验证触发准确性、操作安全性和最终 Vault 状态。

## 实施步骤

1. 为每个 Skill 定义 trigger、non-trigger、success、conflict、failure 用例。
2. 用临时 Vault 和共享 fixture 执行 CLI 调用。
3. 保存期望命令计划和最终文件摘要。
4. 验证写入范围没有超出 Skill 声明。
5. 验证失败和冲突时 Vault 保持预期状态。
6. 增加危险 shell 内容、Unicode 和多行 Markdown 用例。
7. 在 CI 中运行确定性评测；模型主观质量评测单独记录。
8. 生成 Skill capability/version 兼容矩阵。

## 交付物

- `skills/evals/`
- 评测运行器
- 兼容矩阵和发布报告

## 验收标准

- [ ] 每个 Skill 至少覆盖五类基本用例。
- [ ] non-trigger 不会执行 Vault 写入。
- [ ] 冲突测试不会静默覆盖。
- [ ] CLI 版本不满足时评测明确失败。
- [ ] CI 能阻止破坏 Skill 契约的发布。

## 验证

```bash
./scripts/lint-skills.sh
./scripts/run-skill-evals.sh
```
