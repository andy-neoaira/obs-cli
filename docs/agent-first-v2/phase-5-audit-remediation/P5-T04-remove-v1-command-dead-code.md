# P5-T04：删除未注册的 V1 命令死代码

- 状态：`已完成`
- 优先级：`中`
- 负责人：`待分配`
- 涉及项目：`obs-cli`
- 依赖：P1-T08
- 来源：审计 m2

## 背景与已确认现象

V2 根命令只注册 Agent-first 命令树，但 `cmd/` 仍保留旧 Human-first Cobra 命令、
包级变量和 `init()`。这些命令不可达，却继续产生初始化、维护和误注册风险。

## 目标

删除未注册的 V1 命令及其专属辅助代码，使 `cmd/` 只保留当前 V2 命令树和仍被引用的
公共实现。

## 非目标

- 不删除迁移指南、许可证归因和历史说明。
- 不因为“看起来旧”而批量删除 `pkg/actions` 或 `pkg/obsidian`。
- 不改变 V2 命令名称、参数和响应。
- 不提供 V1 兼容别名。

## 候选删除范围

- `cmd/add_vault.go`
- `cmd/create.go`
- `cmd/daily.go`
- `cmd/delete.go`
- `cmd/editor.go`
- `cmd/frontmatter.go`
- `cmd/list.go`
- `cmd/list_vaults.go`
- `cmd/move.go`
- `cmd/open.go`
- `cmd/print.go`
- `cmd/remove_vault.go`
- `cmd/search.go`
- `cmd/search_content.go`
- `cmd/set_default.go`
- `cmd/content_input.go`
- 只覆盖上述代码的测试

最终删除集合必须由引用检查决定。

## 执行过程

1. 用 `rg` 和 `go list` 建立旧符号引用清单，区分 V1 专属和共享实现。
2. 先确认根命令协议测试覆盖所有禁止重新出现的旧命令。
3. 删除 V1 Cobra 命令、专属全局变量、`init()` 和专属测试。
4. 运行编译与测试，识别仍被 V2 使用的共享 helper；共享代码不得误删。
5. 对删除后完全无引用的包生成候选清单，但跨包清理另建任务，不在本任务扩大。
6. 更新 V1→V2 迁移文档，明确源码实现已移除但迁移映射仍保留。

## 验收标准

- [x] 根命令仍只暴露 V2 命令和预留 namespace。
- [x] 旧命令名全部返回未知命令，不存在隐式别名。
- [x] `cmd` 不再因旧命令执行 Human-first `init()`。
- [x] V2 内容输入、stdin、多行 Markdown 和 JSON 测试保持通过。
- [x] 未删除任何仍被 V2 引用的实现。
- [x] 二进制、测试和许可证门禁通过。

## 验证命令

```bash
rg -n 'addVaultCmd|createNoteCmd|DailyCmd|searchContentCmd' cmd
go test ./cmd -run 'Root|Protocol|Input|V2'
go test -race ./cmd
make release-check
git diff --check
```

## 回滚

该任务必须是纯删除和必要引用修正的独立提交，可整体 revert。不得与协议行为修改
混在同一个提交。

## 完成记录

- 完成日期：`2026-07-28`
- 删除：16 个未注册 V1 命令/辅助文件及 3 个仅覆盖这些死代码的测试文件。
- 保留：V2 命令、公共业务包、迁移映射和许可证归因均未删除。
- 验证：`go test ./cmd`、根协议测试和 `make release-check` 通过；删除无效测试后
  有效代码总覆盖率由 `72.9%` 提升到 `73.8%`。
