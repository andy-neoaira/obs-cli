# Skill Eval 发布报告

- 最低 CLI：`v1.0.0-rc.1`
- 协议：`obs-cli/v1`
- 场景清单：`scenarios.json`
- 基本场景契约：`55 / 55` 已定义并通过结构/交叉约束校验（11 个 Skill ×
  trigger/non-trigger/success/conflict/failure）
- 跨场景安全用例：3（危险 shell 与 Unicode/多行 Markdown、stale revision、
  低版本/缺失 capability）

## 确定性评测

发布门禁通过 `./scripts/run-skill-evals.sh` 执行：

1. 严格校验所有 Skill 的 V1 契约。
2. 校验场景清单完整覆盖当前 Skill 目录，且每个 Skill 恰有五类基本契约用例。
3. 对照 CLI 实际 capability surface 和 Skill 声明验证兼容矩阵。
4. 在临时 Vault 中验证 dry-run 无写入、危险内容按原始 bytes 保存。
5. 注入外部修改后以 stale revision 执行 append，验证
   `REVISION_CONFLICT` 且 Vault digest 不变。
6. 验证低于最低版本和缺失 capability 会明确失败。

以上契约检查和跨场景 CLI 安全用例完全确定、无需网络或模型调用，属于 CI 与
release-check 的阻断门禁。55 个 prompt 不声称已经由模型逐条执行；当前仓库没有
内置模型路由器，trigger/non-trigger 的实际选择质量归入下述独立模型评测。

## 模型主观质量评测

触发边界、报告文风、综合质量和建议价值需要使用固定提示集做独立模型评审。该部分
记录在后续发布候选评审中，不伪装成确定性单元测试，也不影响安全契约门禁；任何模型
评审不得获得超出 `scenarios.json` 声明的写入权限。
