# P5-T01：配置原子替换与目录持久化

- 状态：`已完成`
- 优先级：`高`
- 负责人：`待分配`
- 涉及项目：`obs-cli`
- 依赖：P1-T03
- 来源：审计 M1

## 背景与已确认现象

`pkg/config/v2_store.go` 自行执行临时文件写入、文件 sync、close 和 rename，
但 rename 后未同步父目录。配置写入也没有复用 `pkg/storage` 已实现的跨平台
replace 和目录同步语义。

## 目标

配置更新在保留现有配置锁、校验和 `0600` 权限的前提下，完成可测试的
“文件同步 + 原子替换 + 父目录同步”提交。

## 非目标

- 不给 revision 存储 API 增加无条件覆盖开关。
- 不改变 `config-v2.json` 格式、路径、Vault ID 或锁超时语义。
- 不修改 Note 写入协议和错误 envelope。

## 修改范围

- `pkg/config/v2_store.go`
- `pkg/config/v2_store_test.go`
- `pkg/storage/atomic.go` 及平台文件（仅在抽取通用原子替换原语时）
- `pkg/storage/*_test.go`
- P1/P5 相关存储文档

## 执行过程

1. 为配置提交增加可注入检查点，复现 rename 后目录同步失败。
2. 选择并记录实现：
   - 首选：从 storage 抽取不包含业务 precondition 的底层原子替换原语；
   - 备选：配置层补齐与 storage 相同的跨平台 replace/dir-sync。
3. 保留 `Store.Update` 的配置锁、读改写、Schema 校验和 `0600` 权限。
4. rename 成功后同步配置文件父目录；Windows 使用明确的兼容实现。
5. 验证任一提交前失败不会破坏旧配置或留下临时文件。
6. 验证提交后目录 sync 失败会返回错误，且不会谎报成功。
7. 更新 P1-T03 完成记录，说明配置存储已纳入共同持久性语义。

## 验收标准

- [x] `pkg/config` 不再直接执行缺少目录同步的裸 `os.Rename`。
- [x] 成功写入后的文件权限保持 `0600`。
- [x] 并发配置更新仍由现有配置锁串行化。
- [x] write、file sync、close、replace、directory sync 失败均有测试。
- [x] 失败不会产生截断配置或遗留原子替换临时文件。
- [x] macOS 测试与 Linux/Windows 交叉构建通过。

## 验证命令

```bash
go test ./pkg/config ./pkg/storage -run 'Atomic|Config|Failure|Sync'
go test -race ./pkg/config ./pkg/storage
make build-check
make release-check
git diff --check
```

## 回滚

单独回滚本任务提交，恢复原配置写入实现；回滚后必须重新执行
`go test ./pkg/config` 和 `make release-check`。

## 完成记录

- 完成日期：`2026-07-28`
- 实现：新增 caller-serialized `storage.Store.ReplaceAtomic`，配置层保留原有锁和
  Schema 校验并复用跨平台 replace、父目录同步及写后校验。
- 失败覆盖：部分写入、flush/close 前后、replace 前、directory sync 前后及
  post-commit 验证。
- 验证：定向测试、目标包 race、六目标交叉构建和 `make release-check` 通过；
  总覆盖率 `72.8%`。
