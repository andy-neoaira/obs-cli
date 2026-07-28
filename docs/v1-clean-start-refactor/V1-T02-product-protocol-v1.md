# V1-T02：重置产品版本和 CLI 协议

- 状态：`完成`
- 优先级：`阻断`
- 涉及项目：`obs-cli`
- 依赖：V1-T01

## 目标

把未发布的 Agent-first 工作线一次性重置为产品 `v1.0.0-rc.1` 和唯一线协议
`obs-cli/v1`，并让运行时、capabilities、Schema 基线、fixture 和发布变量使用同一值。

## 修改范围

- `pkg/protocol/protocol.go`
- `cmd/capabilities.go`
- `cmd/root.go`
- `scripts/rc-smoke.sh`
- `skills/evals/scenarios.json`
- `skills/evals/scenarios.schema.json`
- `skills/evals/compatibility-matrix.md`
- `docs/compatibility.json`
- 协议 golden fixture
- 协议和版本相关测试

Schema 文件重命名由 T05 完成；本任务可以先修改内容常量，但必须和 T05 在同一集成分支
合并后才允许对外使用。

## 禁止事项

- 不同时声明 `obs-cli/v1` 和 `obs-cli/v2`。
- 不在 capabilities 中保留旧协议列表项。
- 不接受旧 envelope。
- 不通过环境变量开启兼容模式。
- 不把产品 SemVer 和 protocol version 绑定成同一常量。
- 不改变 `obs-write/v1` 或 `vault-contract/v1`。

## 执行步骤

1. 增加/修改协议测试，先断言唯一协议为 `obs-cli/v1`。
2. 将 `protocol.Version` 改为 `obs-cli/v1`。
3. 更新 capabilities 的 `protocol_versions` 和错误 details。
4. 更新协议 golden JSON 的 `protocol_version`。
5. 将 RC smoke 默认版本改为 `v1.0.0-rc.1`。
6. 更新 Skill eval 的最低 CLI 版本和协议要求。
7. 更新 compatibility matrix：
   - candidate version；
   - minimum CLI version；
   - compatible set ID；
   - CLI version range；
   - required protocol。
8. 验证普通构建仍由 ldflags/BuildInfo 解析产品版本，不在源码恢复硬编码版本。
9. 增加负向测试：`obs-cli/v2` envelope 必须被消费者视为不兼容。

## 测试矩阵

| 场景 | 预期 |
|---|---|
| 成功 envelope | `protocol_version == obs-cli/v1` |
| 失败 envelope | 同样使用 `obs-cli/v1` |
| capabilities | 只包含 `obs-cli/v1` |
| RC smoke | `cli_version == v1.0.0-rc.1` |
| Skill 最低版本 | `v0.x` 失败，`v1.0.0-rc.1` 通过 |
| 旧协议输入 | 明确判定不兼容，不 fallback |

## 验收标准

- [x] 运行时只输出 `obs-cli/v1`。
- [x] capabilities 只声明 `obs-cli/v1`。
- [x] 产品候选版本为 `v1.0.0-rc.1`。
- [x] 所有协议 fixture 与运行时一致。
- [x] `obs-write/v1` 和 `vault-contract/v1` 未变化。
- [x] 不存在双协议代码路径。

## 验证命令

```bash
go test ./pkg/protocol ./cmd
make compatibility-check
make skill-evals
make rc-smoke
rg -n 'obs-cli/v2|v2\.0\.0-rc\.1' cmd pkg scripts skills testdata docs/compatibility.json
git diff --check
```

最后一个 `rg` 必须无输出；第三方依赖不在本任务检索范围。

## 回滚

按单个提交回滚。回滚后不得继续执行 T07；如果 T05 已合并，必须同时回滚 Schema 和
fixture 重命名提交，恢复一个内部一致的状态。

## 完成记录

- 完成日期：`2026-07-28`
- 提交：`bb2a823`
- 定向测试：`make release-check 与候选二进制 RC smoke 通过`
- 协议样例：`protocol_version == "obs-cli/v1"`，且
  `protocol_versions == ["obs-cli/v1"]`
