# P1-T02：安全路径解析

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli`
- 依赖：P0-T02、P0-T05

## 目标

建立唯一的 Vault 路径解析器，阻止 `..`、绝对路径和符号链接逃逸。

## 实施步骤

1. 新建独立 path policy 包，禁止业务层自行拼接路径。
2. 对 Vault 根目录和现有目标执行 canonicalize。
3. 对新文件解析最近存在父目录的真实路径。
4. 处理 macOS 大小写不敏感和 Windows volume/path separator。
5. 定义允许与禁止的隐藏目录。
6. 为 note、附件、模板、Daily Note 接入同一解析器。
7. 使用 P0-T05 fixture 补齐攻击与边界测试。

## 交付物

- 统一安全路径包
- 业务层迁移
- 路径安全测试

## 验收标准

- [x] `../outside.md` 返回 `PATH_OUTSIDE_VAULT`。
- [x] Vault 内指向外部的符号链接被拒绝。
- [x] 合法 Unicode、空格和嵌套路径可用。
- [x] 不存在的新文件也能安全验证父目录。
- [x] 所有写操作均无法绕过解析器。

## 验证

```bash
go test ./... -run 'Path|Symlink|Traversal|Canonical'
```

## 完成记录

- 完成日期：`2026-07-24`
- 新增独立 `pkg/pathpolicy`，使用真实 Vault 根、路径段语义和最近存在父目录解析，不再依赖字符串前缀。
- 统一拒绝 `..`、绝对路径、盘符、UNC、ADS、NUL、隐藏路径和 Vault 外符号链接，错误保留 `PATH_OUTSIDE_VAULT` 与对应 VC rule。
- Vault 内符号链接允许读取，但 Create、Delete、Move、Set、Daily 与链接批量重写拒绝通过物理别名修改。
- Move/Delete 在安全解析前补齐 `.md`，消除了校验目标与实际修改目标不一致的绕过。
- 链接批量重写先预检全部目标，路径冲突不会造成部分写入。
- `go test ./...`、`go test -race ./...`、目标路径测试、`go vet ./...` 与 license check 均通过。
