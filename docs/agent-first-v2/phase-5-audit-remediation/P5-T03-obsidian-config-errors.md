# P5-T03：Obsidian 配置严格读取与错误可见性

- 状态：`已完成`
- 优先级：`高`
- 负责人：`待分配`
- 涉及项目：`obs-cli`
- 依赖：P1-T04、P2-T04
- 来源：审计 m6

## 背景与已确认现象

`ExcludedPaths`、`DefaultNoteFolder` 和 `ReadDailyNotesConfig` 把文件不存在、
权限/I/O 错误及 JSON 损坏都降级为空配置。复核调用链后确认，当前 V2 直接受影响的
是 Daily；V2 `note.move` 使用 `noteops` 结构化重写，不读取 `app.json`。

## 目标

区分“配置不存在”和“配置存在但不可用”，保证 V2 写操作不会静默采用不可信默认值。

## 非目标

- 不要求 Vault 必须存在 `.obsidian/`。
- 不改变合法 Obsidian 配置的 folder、format、template 和 ignore 语义。
- 不把缺少可选配置文件当作错误。
- 不在本任务全面重写旧 `pkg/obsidian` API。

## 修改范围

- `pkg/obsidian/config.go`
- `pkg/obsidian/config_test.go`
- `cmd/daily_v2.go` 及测试
- 协议错误映射和 Daily 规范

## 目标行为

| 情况 | 读取行为 | V2 修改操作 |
|---|---|---|
| 配置文件不存在 | `found=false`，使用文档化默认值 | 允许 |
| 合法配置 | 返回解析结果 | 允许 |
| JSON 损坏 | 返回带文件上下文的错误 | 拒绝 |
| 权限或 I/O 错误 | 返回带文件上下文的错误 | 拒绝 |
| 字段值不支持 | 返回 `INVALID_ARGUMENT` | 拒绝 |

## 执行过程

1. 引入严格读取接口，返回配置、是否存在和错误；只对 `os.ErrNotExist` 安全降级。
2. 保留旧宽松包装器时必须标记其适用范围，V2 修改路径不得继续调用宽松接口。
3. `daily get/create/append` 使用严格 Daily 配置读取。
4. Daily 配置损坏或 format 不支持时返回结构化 `INVALID_ARGUMENT`，包含
   `config_file`，不得退回默认日期路径。
5. 记录 `app.json` 严格 loader；其宽松 V1 调用方由 P5-T04 删除，不把未使用该
   配置的 V2 Move 错误耦合到 `app.json`。
6. 对纯读取场景如采用 warning，warning 必须出现在 JSON data 中并有 Schema/测试；
   未设计稳定 warning 前优先返回明确错误。
7. 增加不存在、合法、损坏、权限错误和不支持字段测试。

## 验收标准

- [x] 缺少配置文件仍按当前文档默认值工作。
- [x] 损坏 `daily-notes.json` 时不会创建或修改任何 Daily Note。
- [x] `app.json` 提供严格 loader；V2 Move 不增加虚假的配置依赖。
- [x] 错误 envelope 可区分配置文件、错误类别和受影响 operation。
- [x] 合法配置的现有 Daily、template、exclude fixture 全部通过。
- [x] 错误 details 只返回 Vault 内配置相对路径，不泄露 Vault 绝对路径。

## 验证命令

```bash
go test ./pkg/obsidian ./pkg/noteops ./cmd -run 'Config|Daily|Move|Excluded'
go test -race ./pkg/obsidian ./pkg/noteops ./cmd
./scripts/run-three-client-e2e.sh
make release-check
git diff --check
```

## 回滚

严格读取 API 和调用方必须在同一提交中回滚，避免出现编译通过但 V2 又静默降级的
中间状态。

## 完成记录

- 完成日期：`2026-07-28`
- 实现：新增 `LoadAppConfig`、`LoadDailyNotesConfig` 和带 file/kind 的
  `ConfigFileError`；仅 `os.ErrNotExist` 被视为可选配置缺失。
- V2 行为：`daily.get/create/append` 对损坏或不可读配置返回
  `INVALID_ARGUMENT`，不创建默认路径 Note。
- 边界修正：确认 V2 `note.move` 不消费 `app.json`，未引入与实际实现不符的阻塞。
- 验证：配置/Daily 定向测试、`pkg/obsidian`/`cmd` race 和
  `make release-check` 通过；总覆盖率 `72.9%`。
