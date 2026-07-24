# P1-T08：V2 命令树、质量门禁与发布迁移

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli`
- 依赖：P1-T01 至 P1-T07、P0-T06

## 目标

完成破坏性 V2 收口，删除 Agent 核心中的 Human-first 遗留并建立发布门禁。

## 实施步骤

1. 收敛为 `capabilities/vault/note/search/metadata/link/daily/template/batch/doctor`。
2. 删除或隔离 picker、编辑器启动、Obsidian URI 和 TTY 确认。
3. 更新 README、命令参考、迁移指南和示例。
4. 增加 gofmt、vet、test、race、coverage、protocol schema 检查。
5. 增加跨平台构建。
6. 执行许可证发布检查。
7. 创建 V2 release candidate 并运行 smoke tests。
8. 记录已知限制和回滚方式。

## 交付物

- V2 命令树
- `docs/migration/V1_TO_V2.md`
- CI 和发布配置
- V2 RC 验证记录

## 验收标准

- [x] 无未文档化的兼容别名或隐式行为。
- [x] README 示例全部可以运行。
- [x] `gofmt`、test、race、vet 和 release-check 通过。
- [x] 二进制归档包含许可证与第三方声明。
- [x] V2 RC 可在临时 Vault 完成完整 CRUD 与冲突测试。

## 验证

```bash
gofmt -l .
go vet ./...
go test ./...
go test -race ./...
make release-check
```

## 完成记录

- V2 根命令收口为 `capabilities/vault/note` 与七个明确返回 `CAPABILITY_UNSUPPORTED` 的预留命名空间。
- V1 picker、GUI/编辑器启动、Obsidian URI、TTY 和隐式 Vault 命令解除根注册；无兼容别名。
- README 中英文版已改为可执行 V2 示例，并新增命令参考与 V1→V2 逐命令迁移指南。
- `release-check` 统一执行 gofmt、vet、test、race、70% coverage、Schema、六目标交叉构建、license 和 RC smoke。
- CI 与 tag release 在发布前执行相同门禁；GoReleaser 归档包含项目及 vendored 许可证/NOTICE。
- 本地 `v2.0.0-rc.1` 临时候选二进制通过 capabilities 版本校验和临时 Vault CRUD/冲突 smoke；未自动创建或推送 tag。
- 已知限制和回滚方式记录在 `docs/migration/V1_TO_V2.md`，验证证据见 `docs/releases/V2_RC_VALIDATION.md`。
