# obs-cli V2 Note 原子操作

- CLI 协议：`obs-cli/v2`
- 写入协议：`obs-write/v1`
- 状态：P1-T06 已实现

## 1. 命令面

```text
obs note list
obs note get <path>
obs note create <path> --content-file <file|->
obs note append <path> --content-file <file|-> [--section <heading>] [--if-match <revision>]
obs note patch <path> --match-file <file|-> --content-file <file|-> --if-match <revision>
obs note replace <path> --content-file <file|-> --if-match <revision>
obs note delete <path> --if-match <revision>
obs note move <source> <target> --if-match <revision>
```

所有命令使用 `--output json`、`--request-id` 和 `--vault`。所有修改命令支持 `--dry-run`。路径是 Vault 内逻辑路径；省略 `.md` 时 CLI 自动补充。绝对路径、父目录跳转、隐藏路径和写入符号链接别名均被拒绝。

## 2. 读取

`note.get` 从一次稳定文件快照返回：

- 原始 `content`，不转换 BOM、换行或 Unicode；
- 解析后的 `frontmatter` object；没有 Frontmatter 时为空 object；
- Vault 逻辑 `path`；
- 基于原始字节的 `revision`。

Frontmatter 存在但 YAML 无效时返回 `INVALID_FRONTMATTER`。`note.list` 返回按逻辑路径排序的 Markdown 文件，不遍历隐藏目录。

## 3. 内容输入

修改正文只能使用 `--content-file`；patch 的匹配上下文使用 `--match-file`。值为 `-` 时从 stdin 读取，因此多行 Markdown 不进入 shell 参数、进程列表或命令历史。

一个命令最多有一个 stdin 消费者。patch 的 match 和 replacement 不能同时使用 `-`，但其中一个可以来自 stdin，另一个来自权限受调用方控制的文件。

## 4. 写入语义

### 4.1 Create

create 使用 `must_not_exist`，目标存在时绝不覆盖。冲突响应的 details 包含：

- `existing_revision`
- `requested_revision`
- `same_content`

这使 Agent 能在超时重试后识别“目标已经具有请求内容”，但响应仍为 `ALREADY_EXISTS`。request ID 只用于关联，不自动等价于幂等键；调用方不得把任意同 request ID 的写入视为已应用。

### 4.2 Append

append 先读取稳定快照，再以该 revision 执行原子替换，不使用裸 `O_APPEND`。

- 普通追加：当原内容和新增内容均非空且原内容末尾不是换行时，在边界补一个 `LF`；不改写其他换行。
- `--section`：按大小写敏感的完整 ATX 标题文本匹配，将内容插入该标题区域末尾、下一个同级或更高级标题之前。
- section 不存在或不唯一时不写入并返回冲突/歧义错误。
- 可选 `--if-match` 用于把调用方先前读取的 revision 加入前置条件。

### 4.3 Patch

patch 对原始字节执行唯一上下文替换：

- `--if-match` 必填；
- match 必须非空且恰好出现一次；
- 0 次匹配返回 `REVISION_CONFLICT`；
- 多次匹配返回 `AMBIGUOUS_NOTE`；
- 任一失败都保持目标原始字节不变。

### 4.4 Replace 与 Delete

replace/delete 默认要求格式严格的 `--if-match sha256:<64 lowercase hex>`。只有人类明确传入 `--unsafe-no-if-match` 时，CLI 才会用操作前即时读取的 revision，并在 dry-run plan 中报告风险。默认 Skills 禁止使用此危险选项。

delete 通过原子存储层先创建 CLI 私有恢复副本，再删除目标；不会把恢复文件写进 Vault。

## 5. Dry-run

dry-run 会完成 Vault 选择、路径解析、内容读取、目标快照、revision 和 patch/section 上下文校验，返回：

- `dry_run: true`
- `applied: false`
- `changed`
- `plan.changes[]`
- `plan.risks[]`
- `plan.preconditions[]`

它不会创建 Note 目录、Note、锁、临时文件或恢复副本。apply 是独立请求，必须重新校验所有前置条件。

## 6. Agent 推荐闭环

1. `capabilities --require note.get --require note.patch`
2. `note get <path>`，保存 `content` 与 `revision`
3. 在内存中分析并生成唯一、足够具体的上下文
4. `note patch ... --if-match <revision> --dry-run`
5. 检查 plan 后执行相同参数的不带 dry-run 请求
6. 再次 `note get`，用 `revision_after` 验证结果

遇到 `REVISION_CONFLICT` 时必须重新读取和重新分析，不能去掉 `--if-match` 重试。

Move 与链接重写的事务语义见 [MOVE_TRANSACTIONS.md](./MOVE_TRANSACTIONS.md)。
