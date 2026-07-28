# V1-T01：冻结命名规范和映射表

- 状态：`实现完成，待提交`
- 优先级：`阻断`
- 涉及项目：`obs-cli`、`miniobsidian.nvim`
- 依赖：无
- 交付物：[NAMING_CONVENTIONS.md](./NAMING_CONVENTIONS.md)

## 目标

在修改任何实现前冻结首版产品、协议、配置、Schema、Capability、内部符号和历史文档的
命名规则，使后续任务只执行已批准映射，不再临时创造名称。

## 硬约束

- Agent-first 是产品 V1，不是旧产品的 V2。
- 协议只允许 `obs-cli/v1`。
- 不批准任何兼容别名、deprecated alias、双协议或双配置读取。
- 内部实现不携带版本后缀。
- 外部协议与持久化格式必须显式从 V1 开始。
- 第三方 module major version不属于改名范围。

## 执行步骤

1. 以全仓检索结果建立项目自有 V2 名称清单。
2. 将名称按以下类别归档：
   - 产品版本；
   - CLI 线协议；
   - 写入协议与 Vault 契约；
   - 配置文件和 Schema；
   - Go 文件、类型、函数；
   - Capability 与 operation；
   - Skill 报告与 golden fixture；
   - 当前产品文档；
   - 历史执行记录；
   - 第三方依赖。
3. 逐项确认 [NAMING_CONVENTIONS.md](./NAMING_CONVENTIONS.md) 的强制映射。
4. 确认五个 `_v2` Capability 全部删除，以 operation discovery 作为事实源。
5. 确认删除历史 V2 执行文档并依赖 Git 历史，不建立仓库内归档副本。
6. 确认删除 `vault migrate` 和旧 `preferences.json` 导入实现。
7. 记录两个仓库待修改文件和当前 HEAD。
8. 评审通过后把命名规范状态改为 `已冻结`。

## 必须明确的决策

- [ ] `ConfigPath` 作为唯一配置路径函数名。
- [ ] 删除 `vault migrate` 和旧 `preferences.json` 导入。
- [ ] 删除五个 `_v2` feature flag，不提供替代别名。
- [ ] 删除 `docs/agent-first-v2`，不在仓库内归档。
- [ ] 删除 `docs/migration/V1_TO_V2.md`，不提供新旧产品迁移承诺。
- [ ] 发布候选版本固定为 `v1.0.0-rc.1`。

## 验收标准

- [ ] 映射表不存在“稍后决定”或多个候选目标名。
- [ ] 两个仓库对协议、产品版本和保留契约的名称一致。
- [ ] 所有项目自有 V2 名称都有明确处置。
- [ ] 第三方 `/v2` 有明确 allowlist 策略。
- [ ] 配置硬切换和本地开发配置处理方式已写明。
- [ ] 后续任务无需重新讨论命名。

## 验证命令

```bash
rg -n -i '\bv2\b|v2_|_v2|config-v2|agent-first-v2|obs-cli/v2' \
  /Users/andy/github/obs-cli \
  /Users/andy/github/miniobsidian.nvim \
  --glob '!**/vendor/**' --glob '!**/.git/**'

git -C /Users/andy/github/obs-cli status --short
git -C /Users/andy/github/miniobsidian.nvim status --short
```

## 回滚

本任务只冻结文档决策。回滚本提交即可，但 T02 开始后不得单独回滚 T01；需要先停止整条
任务链并回滚所有依赖提交。

## 完成记录

- 完成日期：`2026-07-28（工作区实现）`
- obs-cli 提交：`未提交`
- miniobsidian.nvim 基线：`待填写`
- 最终 Capability 决策：`待填写`
- 历史文档决策：`待填写`
