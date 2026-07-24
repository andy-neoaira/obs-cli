# P1-T03：原子写入与 Revision 内核

- 状态：`未开始`
- 负责人：`待分配`
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

- [ ] 所有写命令不再直接调用 `os.WriteFile` 或裸 `os.Rename`。
- [ ] revision 与规范测试向量一致。
- [ ] 冲突时文件字节保持不变。
- [ ] 中断测试不留下半文件或临时文件。
- [ ] 成功、冲突和失败路径都按 `obs-write/v1` 清理锁、临时文件与 journal。
- [ ] 故障注入覆盖 `CONCURRENCY_AND_WRITES.md` 第 9 节列出的提交与回滚位置。
- [ ] `go test -race` 通过。

## 验证

```bash
go test ./... -run 'Atomic|Revision|Conflict|FailureInjection'
go test -race ./...
```
