# P5-T05：Capability 文档与实现一致性门禁

- 状态：`待开始`
- 优先级：`中`
- 负责人：`待分配`
- 涉及项目：`obs-cli`
- 依赖：P1-T05、P1-T07
- 来源：审计 m3

## 背景与已确认现象

运行时声明 `move_plan_preconditions: true`，但 `docs/spec/CAPABILITIES.md`
未解释该稳定 flag。当前门禁也不能发现未来 capability 与规范文档漂移。

## 目标

补全 flag 语义，并建立自动测试保证所有稳定 operation/feature flag 在实现、
Schema 和规范之间可追踪。

## 非目标

- 不新增、删除或重命名 capability。
- 不改变 `note.move` dry-run 或 apply 响应。
- 不修改协议主版本。

## 修改范围

- `docs/spec/CAPABILITIES.md`
- `cmd/capabilities.go`
- `cmd/capabilities_test.go` 或独立文档一致性测试
- 必要时 `scripts/schema-check.sh`

## 执行过程

1. 文档化 `move_plan_preconditions`：
   - dry-run 返回确定性 `plan_hash`；
   - apply 接受 `--if-plan-hash`；
   - 计划变化返回 `REVISION_CONFLICT`。
2. 枚举运行时所有稳定 feature flag，与规范表格做集合比较。
3. 枚举 operation，确保 Schema 与测试不会接受重复或空名称。
4. 将一致性检查加入现有 `schema-check` 或 `release-check`。
5. 增加负向 fixture，证明遗漏文档 flag 时门禁失败。

## 验收标准

- [ ] `move_plan_preconditions` 有准确、可供 Agent 使用的规范说明。
- [ ] 实现新增未文档化 flag 会导致测试失败。
- [ ] 文档残留已删除 flag 会导致测试失败或明确标记 deprecation。
- [ ] operation/flag 名称唯一且稳定。
- [ ] 不改变现有 capabilities JSON 的兼容字段。

## 验证命令

```bash
go test ./cmd -run 'Capabilities|Documentation'
make schema-check
make release-check
git diff --check
```

## 回滚

文档与自动门禁必须一起回滚；不得只回滚文档而保留会永久失败的集合测试。
