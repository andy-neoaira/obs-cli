# P1-T05：Capabilities、通用参数与 Dry-run

- 状态：`未开始`
- 负责人：`待分配`
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

- [ ] Agent 无需解析 `--help` 即可发现能力。
- [ ] dry-run 前后 Vault 内容摘要完全相同。
- [ ] 不支持的 capability 返回 `CAPABILITY_UNSUPPORTED`。
- [ ] capabilities 输出有 Schema 和黄金测试。
- [ ] feature flag 名称稳定且有文档。

## 验证

```bash
go run . capabilities --output json
go test ./... -run 'Capabilities|DryRun|CommonFlags'
```

