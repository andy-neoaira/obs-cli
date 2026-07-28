# V1-T03：清理 Go 内部 V2 命名

- 状态：`实现完成，待提交`
- 优先级：`高`
- 涉及项目：`obs-cli`
- 依赖：V1-T01、V1-T02

## 目标

删除 Go 源码、测试、文件名和注释中的内部版本后缀，使当前实现以领域名称表达职责，
不保留任何 V2 别名。

## 主要映射

```text
cmd/*_v2.go                 -> cmd/*.go
cmd/v2_namespace.go         -> cmd/namespace.go
new*V2Command               -> new*Command
newV2Namespace              -> newNamespace
renderV2                    -> renderEnvelope
pkg/config/v2_store.go      -> pkg/config/store.go
V2Config                    -> Config
NewV2Config                 -> NewConfig
ValidateV2Config            -> ValidateConfig
pkg/obsidian/v2_registry.go -> pkg/obsidian/registry.go
```

完整映射以 [NAMING_CONVENTIONS.md](./NAMING_CONVENTIONS.md) 为准。

## 修改范围

- `cmd/`
- `pkg/config/`
- `pkg/obsidian/`
- 直接引用这些符号的测试
- 相关 GoDoc 和错误消息

## 禁止事项

- 不使用 type alias 保留旧类型名。
- 不增加 wrapper 函数保留旧构造器。
- 不同时保留新旧文件。
- 不借机改变命令参数、响应结构或业务行为。
- 不批量改动第三方 module import path。
- 不覆盖与本任务无关的用户修改。

## 执行步骤

1. 记录改名前测试基线和文件清单。
2. 先重命名文件，再统一修改声明和调用点。
3. 将 `renderV2` 改为职责名称 `renderEnvelope`。
4. 将命令构造器改为无版本名称。
5. 将配置类型和 store 文件改为通用名称；配置语义变更留给 T04。
6. 将 Registry 文件和测试改名。
7. 清理 Short、Long、错误消息和 GoDoc 中把当前实现称为 V2 的文字。
8. 使用 `gofmt` 格式化所有修改文件。
9. 运行包级测试，确认只是符号级重构。
10. 检索旧符号，确保没有别名和注释残留。

## 验收标准

- [ ] `cmd`、`pkg/config`、`pkg/obsidian` 不存在项目自有 `_v2.go` 文件。
- [ ] 不存在 `V2Config`、`renderV2`、`new*V2Command`。
- [ ] 不存在兼容 alias 或 wrapper。
- [ ] 命令树、operation 名和 JSON 响应未变化。
- [ ] GoDoc 使用当前产品语言，不再叙述 V2 升级。
- [ ] 相关包测试和 race 测试通过。

## 验证命令

```bash
go test ./cmd ./pkg/config ./pkg/obsidian
go test -race ./cmd ./pkg/config ./pkg/obsidian
go vet ./...
find cmd pkg -type f -iname '*v2*' -print
rg -n 'V2Config|NewV2Config|ValidateV2Config|renderV2|new[A-Za-z]+V2Command|newV2Namespace' cmd pkg
git diff --check
```

两个检索命令必须无输出。

## 回滚

文件移动和符号修改必须位于同一提交，以便整体回滚。不得只恢复旧文件名而保留新符号，
或只恢复旧符号而留下新测试。

## 完成记录

- 完成日期：`2026-07-28（工作区实现）`
- 提交：`未提交`
- 重命名文件数：`待填写`
- 删除旧符号数：`待填写`
