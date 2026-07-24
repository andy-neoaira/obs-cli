# P1-T05：Capabilities、通用参数与 Dry-run

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli`
- 依赖：P1-T04

## 目标

允许 Agent 在执行前发现协议、命令和特性，并为所有修改操作提供无副作用计划。

## 实施步骤

1. 实现 `obs capabilities --output json`。
2. 返回 CLI 版本、协议版本、Vault 规范版本和 feature flags。
3. 统一注册 `--output`、`--request-id`、`--dry-run`、`--if-match`、`--vault`。
4. 为每个修改命令建立 plan 对象。
5. `--dry-run` 返回目标、预期变化、风险和前置条件，不写文件。
6. 定义 capability 的弃用与新增策略。
7. 补充 Skill 调用前 capability 检查示例。

## 交付物

- `capabilities` 命令
- 通用参数中间层
- dry-run plan 模型

## 验收标准

- [x] Agent 无需解析 `--help` 即可发现能力。
- [x] dry-run 前后 Vault 内容摘要完全相同。
- [x] 不支持的 capability 返回 `CAPABILITY_UNSUPPORTED`。
- [x] capabilities 输出有 Schema 和黄金测试。
- [x] feature flag 名称稳定且有文档。

## 验证

```bash
go run . capabilities --output json
go test ./... -run 'Capabilities|DryRun|CommonFlags'
```

## 完成记录

- 新增 `capabilities.get`、operation/feature flag 清单及 JSON Schema。
- 建立 `--output`、`--request-id`、`--dry-run`、`--if-match`、`--vault` 公共参数绑定层。
- `vault.add`、`vault.remove`、`vault.set-default`、`vault.migrate` 均提供结构化 dry-run plan。
- 预演路径只读取注册表和 Vault；测试验证不会创建配置、锁或更改 Vault 摘要。
- 稳定名称、演进策略和场景化 Skill 协商流程见 `docs/spec/CAPABILITIES.md`。
