# P1-T03：原子写入与 Revision 内核

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli`
- 依赖：P0-T04、P1-T02

## 目标

提供所有写操作共享的存储层，实现 revision、条件更新和原子替换。

## 实施步骤

1. 实现基于原始字节 SHA-256 的 revision。
2. 读取结果返回内容、元数据和 revision。
3. 实现 `WriteAtomic(path, data, expectedRevision)`。
4. 临时文件必须位于目标同目录并继承合理权限。
5. 写入后 flush、close，再执行原子 rename。
6. revision 不匹配时返回 `REVISION_CONFLICT` 且不修改文件。
7. 对创建、覆盖、删除分别定义前置条件。
8. 注入失败点，测试写入中断与清理。

## 交付物

- 原子存储包
- revision 测试向量
- 故障注入测试

## 验收标准

- [x] 所有写命令不再直接调用 `os.WriteFile` 或裸 `os.Rename`。
- [x] revision 与规范测试向量一致。
- [x] 冲突时文件字节保持不变。
- [x] 中断测试不留下半文件或临时文件。
- [x] 成功、冲突和失败路径都按 `obs-write/v1` 清理锁、临时文件与 journal。
- [x] 故障注入覆盖 `CONCURRENCY_AND_WRITES.md` 第 9 节列出的提交与回滚位置。
- [x] `go test -race` 通过。

## 验证

```bash
go test ./... -run 'Atomic|Revision|Conflict|FailureInjection'
go test -race ./...
```

## 完成记录

- 完成日期：`2026-07-24`
- 新增 `pkg/storage`：原始字节 revision、稳定快照、跨进程锁、条件原子写、可恢复删除、双路径移动和多文件事务。
- Create、Append、Overwrite、Set、Delete、Move、Daily 与链接批量重写已移除业务层直接文件写入。
- P5-T01 将 V2 配置更新纳入共同原子替换语义：保留配置锁与校验，并在跨平台
  replace 后同步父目录、执行写后校验和失败清理。
- 链接重写使用多文件 stage/journal/commit/rollback；回滚失败保留恢复资料并返回 `PARTIAL_FAILURE`。
- 故障注入覆盖临时文件、部分 write、flush/close、提交前复核、提交、提交后验证、多文件第 N 项提交和第 N 项回滚。
- 全量测试、全量 Race、Vet、license check 和 Windows storage 交叉编译通过。
