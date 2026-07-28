# Revision、原子写入与冲突协议

- 协议标识：`obs-write/v1`
- 状态：已实现并通过双仓契约测试
- 发布日期：2026-07-24
- CLI 协议：`obs-cli/v1`
- Vault 规范：`vault-contract/v1`
- 测试向量：[revision-v1.json](../../testdata/revision/revision-v1.json)

## 1. 目标与限制

### CW-1.1 目标

本协议用于降低 Obsidian、Agent、同步服务和 Neovim 同时操作同一 Vault 时的覆盖风险，并保证单文件不会以半写入状态对其他客户端可见。

### CW-1.2 文件系统限制

普通跨平台文件系统不提供“仅当目标内容 hash 等于 X 时原子替换”的通用 compare-and-swap。`obs-write/v1` 采用 revision、协作锁、提交前复核、原子 rename 和恢复数据缩小竞态窗口，但不能对不遵守锁的外部进程提供数学上的全局串行化保证。

实现和 Skill 不得宣称能够阻止所有纳秒级外部竞态。检测到变化时必须停止；无法排除竞态时必须保留可恢复信息。

## 2. Revision

### CW-2.1 算法

文件 revision 为磁盘原始字节的 SHA-256：

```text
sha256:<64 lowercase hexadecimal characters>
```

计算输入包括 BOM、CRLF/LF 和末尾换行。不得在 hash 前解码文本、规范化 Unicode、解析 Frontmatter 或转换换行。

### CW-2.2 不存在状态

不存在的目标没有 revision。create 使用 `must_not_exist: true` 前置条件，不使用空字符串、全零 hash 或 `null` revision 伪装存在状态。

### CW-2.3 读取快照

读取操作应：

1. 以 no-follow 或等价安全方式打开已经通过 Vault 边界验证的文件。
2. 读取文件身份、大小和修改时间。
3. 读取全部原始字节并计算 revision。
4. 再次读取文件身份、大小和修改时间。
5. 前后不一致时丢弃结果并有限重试；仍不稳定则返回 transient read error。

返回的 content 和 revision 必须来自同一次稳定快照。

### CW-2.4 响应

所有读取和成功修改响应必须返回 revision。修改响应分别返回：

- `revision_before`：创建时不存在。
- `revision_after`：删除时不存在。
- `changed`：字节是否变化。

## 3. 操作前置条件

### CW-3.1 Create

create 必须使用 `must_not_exist: true`。目标已经存在时返回 `ALREADY_EXISTS`，即使内容相同也不得静默当作新建成功；具体幂等请求可以返回先前已确认的同一 request 结果。

### CW-3.2 Append

append 不是裸 `O_APPEND`。实现必须读取稳定快照、用明确换行规则构造新字节，并以该 revision 条件执行原子替换。

### CW-3.3 Patch 与 Replace

patch 和 replace 必须携带 expected revision。patch 还必须验证上下文唯一匹配。revision 或 patch 上下文任一不满足时，目标字节保持不变。

### CW-3.4 Move

move 计划必须记录源 revision、目标不存在前置条件以及每个链接更新文件的 revision。apply 前重新验证全部前置条件。

### CW-3.5 Delete

delete 必须携带 expected revision。默认删除进入可恢复隔离区或创建恢复副本；永久删除需要独立 capability 和明确授权。

### CW-3.6 强制写入

绕过 expected revision 的 force 操作不属于默认 `obs-cli/v1` capability，Skills 禁止使用。未来如增加，必须拥有独立 operation、醒目的 plan 风险和审计记录。

## 4. 协作锁

### CW-4.1 锁范围

CLI 写入必须获取基于真实目标路径的进程间排他锁。多文件事务按规范化路径字典序获取锁，避免 CLI 进程互相死锁。

### CW-4.2 锁位置

锁文件默认存储在 CLI 自有运行时目录，不写入普通笔记目录。锁键由 Vault 真实身份和目标真实路径派生，不在文件名中暴露完整用户路径。

### CW-4.3 锁的边界

Obsidian、Neovim 和同步服务不保证遵守 CLI 锁，因此持锁期间仍必须执行提交前 revision 复核。锁只保证协作的 CLI/Agent 写入彼此串行。

### CW-4.4 过期锁

锁实现必须使用操作系统自动释放机制，或记录进程身份并安全判断过期。不得仅按固定时间删除可能仍有效的锁。

## 5. 单文件原子写入

### CW-5.1 准备

写入前必须依次完成：

1. 解析真实 Vault 和安全目标路径。
2. 获取目标锁。
3. 读取稳定快照并验证 expected revision 或 must-not-exist。
4. 在目标同目录创建随机、权限受限的临时文件。

临时文件必须使用 exclusive create，禁止可预测名称和符号链接跟随。

### CW-5.2 临时文件

实现将完整新字节写入临时文件，检查 write/close 错误，并在平台支持时 flush 文件。更新已有文件时应保留其权限位；不得复制符号链接身份。

### CW-5.3 提交前复核

临时文件准备完成后，必须重新读取目标身份和 revision：

- expected revision 不匹配时删除临时文件并返回 `REVISION_CONFLICT`。
- create 目标出现时删除临时文件并返回 `ALREADY_EXISTS`。
- 目标从普通文件变成符号链接、目录或其他类型时按安全错误停止。

### CW-5.4 原子提交

更新操作使用同一文件系统内的原子 replace。create 使用平台 no-replace 原语，目标出现时绝不覆盖。实现不得直接依赖会在部分平台覆盖目标的通用 `rename` 来实现 create。

平台支持时，提交后 flush 父目录。平台不支持目录 flush 时应记录 capability 限制，但不得退化为原地截断写入。

### CW-5.5 提交验证

提交后重新读取目标，计算 `revision_after` 并确认等于预期新字节的 revision。验证失败返回 I/O 或 partial failure，并保留恢复信息。

### CW-5.6 清理

成功、冲突、取消和普通失败都必须删除本次临时文件并释放锁。无法清理时返回 warning，包含 Vault 内安全逻辑标识或匿名 transaction ID，不泄露 Vault 外路径。

## 6. 多文件事务

### CW-6.1 Plan

多文件操作必须先产生不可变 plan，包含 transaction ID、每个 path/action、expected revision、目标不存在条件、预期新 revision 和风险。

### CW-6.2 Stage

apply 按稳定顺序获取全部协作锁，验证全部前置条件，然后为所有更新准备临时文件和恢复副本。任何 stage 失败时不得开始 commit。

### CW-6.3 Commit 与回滚

跨多个普通文件的提交不是文件系统级全局原子。实现必须：

1. 记录可恢复 transaction journal。
2. 按计划顺序提交。
3. 任一提交失败时逆序回滚已提交项。
4. 回滚后验证 revision。
5. 完整回滚成功时返回原始操作失败；回滚不完整时返回 `PARTIAL_FAILURE`。

### CW-6.4 Partial Failure

`PARTIAL_FAILURE` details 必须包含：

- transaction ID。
- `completed`、`failed`、`rolled_back`、`rollback_failed`。
- 每个当前文件的 revision。
- 可执行但不自动执行的 `recovery_actions`。

调用方不得在未读取恢复清单和当前 revision 时重放整个事务。

### CW-6.5 Journal

Journal 和恢复副本默认位于 CLI 私有运行时目录；需要同文件系统原子操作的临时内容位于各目标同目录并使用内部随机名称。Journal 不保存不必要的笔记正文，恢复副本必须限制权限并在确认成功后清理。

## 7. Neovim 与外部修改

### CW-7.1 Agent 修改前

从 Neovim 发起更新时，handoff 必须携带磁盘 revision。buffer 有未保存修改时不得把磁盘 revision 代表为 buffer 内容；必须先保存，或将操作限制为不写入的内存内容分析。

### CW-7.2 Agent 修改后

Agent 返回 revision before/after。Neovim：

- buffer 未修改时，可以运行 `checktime` 并展示变更。
- buffer 已修改时，禁止自动重载，保留 buffer 和磁盘内容并提供三方比较。

### CW-7.3 外部变化

plan 与 apply 之间发生 Obsidian、同步服务或其他编辑器修改时，expected revision 失效，apply 返回 `REVISION_CONFLICT`。Agent 必须重新读取、重新分析和重新授权，不得去掉前置条件重试。

## 8. 幂等与重试

### CW-8.1 Request ID

request ID 用于关联，不能单独证明幂等。实现可以维护有限期请求结果记录；仅当 operation、参数摘要和输入 revision 都一致时，才能返回先前成功结果。

### CW-8.2 安全重试

- 读取操作可以有限重试 transient error。
- revision conflict 只能在重新读取和重新规划后重试。
- partial failure 只能按恢复清单处理。
- 超时后状态未知时，调用方先读取目标 revision 或查询 transaction 状态，不直接重复写入。

## 9. 故障注入与验收

实现测试至少在以下位置注入失败：

1. 临时文件创建前后。
2. 部分 write 后。
3. flush/close 时。
4. 提交前 revision 复核时。
5. rename/replace 时。
6. 提交后验证时。
7. 多文件第 N 项提交时。
8. 回滚第 N 项时。

每个故障点验证目标文件是旧完整字节或新完整字节，禁止出现部分新内容；并验证临时文件、锁和 journal 的预期生命周期。

## 10. 已知残余风险

- 不协作的外部编辑器可能在最终复核与原子 rename 的极短窗口内写入。
- 某些网络文件系统、云盘虚拟文件系统和同步驱动不保证本地文件系统相同的 rename/flush 语义。
- 移动端同步可能产生独立冲突副本，而不是 revision conflict。

实现必须通过 capability 或 warning 暴露已知文件系统限制。重要批量修改前应建议用户使用 Vault 级快照或同步服务版本历史，但不得把备份建议替代本协议的安全检查。
