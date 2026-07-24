# obs-cli 原子存储实现

- 协议：`obs-write/v1`
- 实现包：`pkg/storage`
- Revision：磁盘原始字节的 `sha256:<lowercase hex>`

## 单文件操作

`ReadSnapshot` 在同一文件描述符上读取前后元数据，并复核路径身份；变化时有限重试。返回原始字节、权限、大小和 revision。

`WriteAtomic` 要求且只接受一种前置条件：

- `must_not_exist`：创建，使用 no-replace 提交；
- `expected_revision`：替换，revision 不一致返回 `REVISION_CONFLICT`。

写入流程为：排序锁定、稳定快照、前置条件检查、同目录随机临时文件、权限继承、write、flush、close、提交前复核、原子提交、目录 flush 和提交后 revision 验证。

`DeleteAtomic` 在 CLI 私有 runtime 中创建权限为 `0600` 的恢复副本后删除。`MoveAtomic` 同时锁定源和目标，先 no-replace 创建完整目标，再删除源；源删除前失败会移除目标。

## 多文件事务

`ApplyTransaction`：

1. 按规范路径排序并获取全部锁；
2. 验证全部前置条件；
3. stage 所有临时文件和恢复副本；
4. 写入 CLI 私有 journal；
5. 按顺序提交；
6. 失败时逆序回滚；
7. 成功或完整回滚后清理临时文件、恢复副本与 journal。

回滚不完整时返回 `PARTIAL_FAILURE` 并保留 journal 和恢复副本，禁止调用方盲目重放事务。

## 平台行为

- Unix 使用 `flock`、同文件系统 `rename` 和目录 `fsync`。
- Windows 使用 `LockFileEx` 和带 `MOVEFILE_REPLACE_EXISTING | MOVEFILE_WRITE_THROUGH` 的 `MoveFileEx`。
- create 使用硬链接 no-replace 提交；不支持该原语的文件系统会明确失败，不退化为覆盖式 rename。

锁文件名由规范目标路径 hash 生成，不泄露 Vault 路径。锁文件可以保留，但操作系统锁会在进程退出时自动释放。
