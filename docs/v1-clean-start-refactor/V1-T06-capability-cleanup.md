# V1-T06：清理 Capability 版本后缀

- 状态：`实现完成，待提交`
- 优先级：`高`
- 涉及项目：`obs-cli`
- 依赖：V1-T01、V1-T02

## 目标

移除 Capability feature flag 中重复的 `_v2` 历史标记，以 operation discovery 作为能力
事实源；只有 operation 无法表达的跨操作特性才保留无版本 feature flag。

## 当前候选

```text
daily_notes_v2
link_inspection_v2
metadata_v2
note_operations_v2
search_v2
```

T01 已冻结为全部删除，不提供改名后的替代 flag。

## 设计规则

1. operation 列表是具体能力的首要事实源。
2. feature flag 只描述跨 operation、无法由名称推断的行为保证。
3. flag 名不携带产品或协议版本。
4. 不保留旧 flag alias。
5. 新客户端必须忽略未知可选 flag，但不能依赖已删除 flag。

## 修改范围

- `cmd/capabilities.go`
- `cmd/capabilities_test.go`
- `docs/spec/CAPABILITIES.md`
- `docs/spec/schema/capabilities-v1.schema.json`
- capability 文档/实现双向一致性测试
- Skill 和 Adapter 中可能存在的 flag 引用

## 执行步骤

1. 对每个 `_v2` flag 判断 operation 是否已完全表达能力。
2. 删除全部五个 `_v2` flag。
3. 不新增领域级替代 flag。
4. 更新文档表格和 operation discovery 说明。
5. 更新实现与文档双向集合测试。
6. 增加负向测试：旧 `_v2` flag 不得出现。
7. 检查 Skills 是否只使用 `capabilities --require <operation>`。
8. 检查 `miniobsidian.nvim` 是否只使用 operation discovery；如有依赖，列入 T07。

## 强制目标

直接删除：

```text
note_operations_v2
metadata_v2
search_v2
link_inspection_v2
daily_notes_v2
```

保留已经表达真实保证的无版本 flag，例如：

```text
atomic_writes
dry_run_plans
json_error_envelopes
multi_file_transactions
move_plan_preconditions
revision_preconditions
vault_discovery_read_only
vault_path_policy
```

不得在本任务中重新引入同义 flag。

## 验收标准

- [ ] Capability 不包含 `_v2` flag。
- [ ] operation 列表继续完整声明实际命令。
- [ ] Skills 通过 `--require operation` 协商能力。
- [ ] 文档与运行时 flag 集合双向一致。
- [ ] 不存在旧 flag alias。
- [ ] capability Schema 和 golden 输出通过。

## 验证命令

```bash
go test ./cmd -run 'Capabilities|Capability'
make compatibility-check
make schema-check
rg -n 'daily_notes_v2|link_inspection_v2|metadata_v2|note_operations_v2|search_v2' \
  cmd pkg skills scripts testdata docs/spec
git diff --check
```

检索必须无输出。

## 回滚

整体回滚 capability 提交。如果 T07 已完成，必须同步回滚 Adapter 配对提交；不得让插件
依赖 CLI 已删除或尚未提供的 flag。

## 完成记录

- 完成日期：`2026-07-28（工作区实现）`
- 提交：`未提交`
- 删除 flag：`待填写`
- 保留/改名 flag：`待填写`
