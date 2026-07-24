# V2 Release Candidate 验证记录

- 候选版本：`v2.0.0-rc.1`（本地候选，不创建或推送 Git tag）
- 验证日期：2026-07-25
- 平台：由 CI 和本地 `build-check` 覆盖
- 发布门禁：`make release-check`
- RC 二进制：`rc-smoke.sh` 以 ldflags 注入 `v2.0.0-rc.1`，验证 capabilities 返回相同版本；产物位于临时目录并在验证后清理

## 自动验证范围

- 非 vendor Go 源码 gofmt
- `go vet ./...`
- 全量单元/集成测试
- `go test -race ./...`
- 总 statement coverage 不低于 70%
- response/capabilities JSON Schema 合同测试
- darwin/linux/windows × amd64/arm64 交叉构建
- 项目、上游和 vendored 依赖许可证检查
- 发布归档配置包含许可证与第三方声明
- 临时 Vault RC smoke

## 临时 Vault smoke

`TestV2RCSmoke` 每次创建隔离的临时 Vault 和 V2 注册表，验证：

1. create 成功且重复 create 返回 `ALREADY_EXISTS`
2. get 返回 content/revision
3. append dry-run 不写入
4. append 使用正确 revision 成功
5. stale replace 返回 `REVISION_CONFLICT`
6. 使用新 revision replace 成功
7. 条件 delete 成功且目标不存在

Move 的目标防覆盖、链接 rewrite、计划后外部修改、完整回滚和 partial failure 恢复清单由 P1-T07 集成/故障注入测试覆盖。

## 归档核对

GoReleaser 配置要求每个归档包含：

- `obs-cli` 二进制
- `LICENSE`
- `THIRD_PARTY_NOTICES.md`
- `vendor/**/LICENSE*`
- `vendor/**/COPYING`
- `vendor/**/NOTICE`

## 已知限制与回滚

已知限制和回滚步骤记录在 [V1 到 V2 迁移指南](../migration/V1_TO_V2.md)。正式创建 tag 前必须在干净工作树重新运行 `make release-check`；本任务不自动创建或推送 release/tag。
