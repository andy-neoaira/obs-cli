# P3-T08：Skill 场景评测与发布门禁

- 状态：`已完成`
- 负责人：`Codex`
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

- [x] 每个 Skill 至少覆盖五类基本用例。
- [x] non-trigger 不会执行 Vault 写入。
- [x] 冲突测试不会静默覆盖。
- [x] CLI 版本不满足时评测明确失败。
- [x] CI 能阻止破坏 Skill 契约的发布。

## 验证

```bash
./scripts/lint-skills.sh
./scripts/run-skill-evals.sh
SKILL_EVAL_CLI_VERSION=v2.0.0-rc.1 make release-check
```

## 完成记录

- 新增 `skills/evals/scenarios.json` 与 Draft 2020-12 Schema，覆盖 11 个 Skill
  的 trigger/non-trigger/success/conflict/failure 共 55 个场景契约。
- 场景清单记录预期选择、operation 计划、写入目标、outcome、Vault 状态和正式
  protocol error code；清单与实际 capability surface、Skill 声明双向校验。
- `write_capabilities` 与 CLI `mutating` 标记精确交叉验证；只读 success 必须保持
  Vault 不变，修改型 success 必须声明实际写 operation 和目标摘要。
- 新增临时 Vault 执行评测：dry-run digest 不变、危险 shell 内容无 sentinel
  副作用、Unicode/多行 Markdown 原样落盘、只读不写、stale revision 与
  ambiguous patch 均不改变 Vault。
- 新增最低版本检查；`v1.9.9` 明确失败，`v2.0.0-rc.1` 通过。GitHub tag release
  和手工 `make release` 都会在任何发布写操作前传入真实版本并运行完整门禁。
- 新增 capability/version 兼容矩阵与发布报告。仓库没有内置模型路由器，因此
  55 个 prompt 是确定性契约覆盖；模型 trigger/报告质量明确作为独立主观评测，
  不伪装成已执行的单元测试。
- `skill-evals` 已接入 PR CI、release-check 与 release workflow。
- 独立前向审计最终 PASS；完整 RC release-check 通过，覆盖率 72.3%。
