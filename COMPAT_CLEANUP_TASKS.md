# 兼容遗留清理执行台账

> 目标：项目尚未正式发布，不保留旧版本兼容入口和隐式兼容行为，统一收敛到最新架构与明确行为。

## 执行任务

- [x] 删除所有 Cobra 命令别名，只保留正式命令名。
- [x] 删除 `delete` 命令残留的无效 `--open/-o` flag。
- [x] 删除 `SearchNotesContent` 旧兼容入口，统一使用 `SearchNotesContentWithOptions`。
- [x] 删除 `Note.UpdateLinks` 兼容委托方法，链接重写只通过 `LinkRewriter`。
- [x] 删除 `link_rewriter.go` 中兼容旧调用的包级函数。
- [x] 收紧笔记路径解析，移除裸文件名 basename 全库回退查找。
- [x] 收紧 vault 输入模型，运行时 vault 参数只接受 vault name，路径型操作使用显式命令语义。
- [x] 删除 `NormalizeContent` 自动转义还原，`--content` 保持字面量，多行内容使用 `--content-file`。
- [x] 创建已有笔记且未指定 `--append/--overwrite` 时返回明确错误。
- [x] 配置文件存在但解析失败时返回错误，不再静默回退。
- [x] `open --editor --section` 改为参数冲突错误。
- [x] URI 参数构造只忽略空字符串，不再吞掉字符串 `"false"`。
- [x] 清理旧项目报告文档和异常 `docs/usage.png` 资产。
- [x] 同步 README、README_CN、skills、测试用例。
- [x] 运行 `go test ./...` 并根据结果修复。
- [x] 完成后重新扫描确认无兼容残留。

## 验证记录

- `GOCACHE=/private/tmp/obs-cli-gocache go test ./...`：通过。
- 兼容残留关键词扫描：未发现旧命令别名、旧 API 包装、旧项目名、旧路径兼容说明残留。
- 扫描中剩余的 `alias` 命中均为 Obsidian wikilink 的正常业务语法（例如 `[[note|alias]]`），不是 CLI 命令别名。
