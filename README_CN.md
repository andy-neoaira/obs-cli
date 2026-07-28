# obs-cli

面向 Agent、非交互、机器可读的 Obsidian Vault 操作 CLI。

`obs-cli` V1 是供 AI Agent、场景化 Skill、脚本和编辑器集成调用的本地 Markdown 执行层。它通过 Vault 边界、revision 前置条件、原子写入、dry-run 计划和稳定 JSON 错误降低 Obsidian、Neovim、同步服务与 Agent 并发修改时的覆盖风险。

本项目基于 [Yakitrak/notesmd-cli](https://github.com/Yakitrak/notesmd-cli) 二次开发。许可信息见 [LICENSE](./LICENSE) 和 [THIRD_PARTY_NOTICES.md](./THIRD_PARTY_NOTICES.md)。

## 当前状态

首个 Agent-first 版本使用以下稳定顶层命令树：

```text
capabilities  vault  note  search  metadata
link          daily  template  batch  doctor
```

Agent 必须通过 `capabilities` 发现已实现 operation。预留命名空间返回 `CAPABILITY_UNSUPPORTED`，不会回退到 picker、GUI 或 TTY 行为。

## 构建

需要 `go.mod` 声明的 Go 版本。

```bash
go build -mod=vendor -o obs-cli .
./obs-cli capabilities --output json
```

## Agent 推荐流程

注册 Vault：

```bash
./obs-cli vault add /absolute/path/to/vault \
  --name Personal \
  --set-default \
  --dry-run

./obs-cli vault add /absolute/path/to/vault \
  --name Personal \
  --set-default
```

Skill 写入前协商能力：

```bash
./obs-cli capabilities \
  --output json \
  --require note.get \
  --require note.patch
```

通过 stdin 安全创建多行内容：

```bash
printf '# Project\n' |
  ./obs-cli note create Projects/demo \
    --vault Personal \
    --content-file -
```

读取内容并保存返回的 revision：

```bash
./obs-cli note get Projects/demo \
  --vault Personal \
  --request-id agent-read-001
```

预演并应用唯一上下文 patch：

```bash
printf 'Project' > /tmp/obs-cli-match.txt
printf 'Project Alpha' |
  ./obs-cli note patch Projects/demo \
    --vault Personal \
    --match-file /tmp/obs-cli-match.txt \
    --content-file - \
    --if-match 'sha256:<revision-from-get>' \
    --dry-run

printf 'Project Alpha' |
  ./obs-cli note patch Projects/demo \
    --vault Personal \
    --match-file /tmp/obs-cli-match.txt \
    --content-file - \
    --if-match 'sha256:<revision-from-get>'
```

示例中的占位 revision 必须替换为 `note get` 返回的精确值。

## 已实现的 V1 operation

```text
capabilities

vault discover
vault list
vault get
vault add
vault remove
vault set-default

note list
note get
note create
note append
note patch
note replace
note delete
note move

daily get
daily create
daily append

metadata get
metadata set

search content
link backlinks
```

所有修改操作支持 `--dry-run`。Note revision 格式为 `sha256:<64 lowercase hex>`。`replace`、`delete`、`patch`、`move` 要求 `--if-match`；`replace/delete` 提供显式 `--unsafe-no-if-match` 逃生参数，默认 Skill 禁止使用。

## 协议保证

- stdout 只输出一个 `obs-cli/v1` JSON envelope。
- stderr 仅用于包含 request ID 的诊断。
- V1 内 operation 名称和错误码保持稳定。
- Vault 逻辑路径拒绝目录穿越、隐藏路径和符号链接逃逸。
- create 永不覆盖。
- append 通过 revision-aware 原子替换实现。
- patch 要求上下文恰好匹配一次。
- move 在一个可恢复事务中创建目标、重写解析后的链接并删除源文件。
- dry-run 不创建配置、锁、临时文件、恢复副本或 Vault 文件。

规范：

- [命令参考](./docs/COMMAND_REFERENCE.md)
- [CLI 协议](./docs/spec/CLI_PROTOCOL.md)
- [Capabilities 与 dry-run](./docs/spec/CAPABILITIES.md)
- [Vault 路径策略](./docs/spec/PATH_POLICY.md)
- [Note 原子操作](./docs/spec/NOTE_OPERATIONS.md)
- [Move 事务](./docs/spec/MOVE_TRANSACTIONS.md)
- [并发与写入](./docs/spec/CONCURRENCY_AND_WRITES.md)

旧命令别名、模糊 picker、编辑器启动、Obsidian URI、基于 cwd 的 Vault 选择和 TTY 确认均不属于 Agent-first 命令表面。

## 开发与发布门禁

```bash
make release-check
```

发布门禁覆盖格式、vet、测试、竞态、覆盖率、协议 Schema、跨平台构建、许可证、发布声明以及临时 Vault CRUD/冲突 RC smoke。

发布归档包含二进制、`LICENSE`、`THIRD_PARTY_NOTICES.md` 和 vendored 依赖的许可证/NOTICE。
