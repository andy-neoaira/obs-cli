# obs-cli

面向 Agent、非交互、机器可读的 Obsidian Vault 操作 CLI。

`obs-cli` V1 是供 AI Agent、场景化 Skill、脚本和编辑器集成调用的本地 Markdown 执行层。它通过 Vault 边界、revision 前置条件、原子写入、dry-run 计划和稳定 JSON 错误降低 Obsidian、Neovim、同步服务与 Agent 并发修改时的覆盖风险。

本项目基于 [Yakitrak/notesmd-cli](https://github.com/Yakitrak/notesmd-cli) 二次开发。许可信息见 [LICENSE](./LICENSE) 和 [THIRD_PARTY_NOTICES.md](./THIRD_PARTY_NOTICES.md)。

## 安装

`obs-cli` 与 Agent Skills 独立安装。可执行文件必须位于 Agent 进程继承的
`PATH` 中；Skill 只是工作流指令包，不包含、也不会自动安装 CLI。

### 一键安装

一条命令安装 CLI、相同版本的全部正式 Skills，并执行离线安装审计：

```bash
curl -fsSL \
  https://raw.githubusercontent.com/andy-neoaira/obs-cli/master/scripts/bootstrap.sh |
  bash -s -- --agent codex
```

未传入 `--version` 时，bootstrap 会解析 GitHub 最新 Release，并为 CLI 和全部
正式 Skills 安装同一个准确 tag。只有需要固定版本时才传入
`--version v1.0.0-rc.1`。

首版本支持以下用户级 Skill 宿主：

| `--agent` | 默认 Skill 目录 |
|---|---|
| `codex` | `${CODEX_HOME:-$HOME/.codex}/skills` |
| `claude-code` | `~/.claude/skills` |
| `opencode` | `${XDG_CONFIG_HOME:-$HOME/.config}/opencode/skills` |
| `cursor` | `~/.cursor/skills` |
| `kimi-code` | `${KIMI_CODE_HOME:-$HOME/.kimi-code}/skills` |

安装到其他受支持的 Agent 时只需修改 `--agent`。bootstrap 不会注册、扫描或修改
Vault，也不会静默覆盖已有 CLI/Skills；托管升级必须显式传入 `--force-cli` 和
`--upgrade-skills`。使用 `--dry-run` 可以预览两部分计划。

### 1. 安装 CLI

普通用户推荐从 GitHub Releases 下载并校验预编译二进制。安装器支持 macOS、
Linux 和 Windows shell，默认安装到 `~/.local/bin`：

```bash
curl -fsSLo /tmp/obs-cli-install.sh \
  https://raw.githubusercontent.com/andy-neoaira/obs-cli/master/scripts/install.sh

# 建议执行前先检查脚本。
less /tmp/obs-cli-install.sh
bash /tmp/obs-cli-install.sh
```

也可以固定版本或指定其他用户可写目录：

```bash
bash /tmp/obs-cli-install.sh \
  --version v1.0.0-rc.1 \
  --install-dir "$HOME/.local/bin"
```

安装器会校验 `checksums.txt`，通过 `--version` 和 `capabilities` 验证下载的
二进制；目标已经存在时默认拒绝覆盖，只有显式传入 `--force` 才会替换。使用
`--dry-run` 可以只查看解析后的平台、下载地址和目标路径。

当前仓库仍处于 release candidate 状态。只有对应 tag 和 GitHub Release 资产发布后，
上述下载命令才可使用。开发者可以从当前源码构建：

```bash
go build -mod=vendor -o obs-cli .
mkdir -p "$HOME/.local/bin"
install -m 0755 obs-cli "$HOME/.local/bin/obs-cli"
```

应从实际启动 Agent 的同一环境验证：

```bash
obs-cli capabilities --output json
```

如果交互式终端能够运行，而 Agent 中找不到命令，应为 Agent 进程配置安装目录，而
不是假定它继承了终端的 `PATH`。

### 2. 单独安装 Agent Skills

正式可分发 Skill 清单位于
[`skills/install-manifest.txt`](./skills/install-manifest.txt)，其中只包含 11 个
`obsidian-*` Skill；`_template` 和 `evals` 是开发资源，永远不会被安装。

下载安装器后，未指定版本时会安装 GitHub 最新 Release 的 Skills：

```bash
curl -fsSLo /tmp/obs-cli-install-skills.sh \
  https://raw.githubusercontent.com/andy-neoaira/obs-cli/master/scripts/install-skills.sh

# 建议执行前先检查脚本。
less /tmp/obs-cli-install-skills.sh
bash /tmp/obs-cli-install-skills.sh --agent codex
```

将 `codex` 替换为 `claude-code`、`opencode`、`cursor` 或 `kimi-code` 即可安装到
其他受支持的宿主。如果新 Skill 未立即出现，请启动新的 Agent 会话。安装器在普通
模式下不会覆盖任何已有目录；显式升级还会校验托管 metadata 和已安装内容的
digest。

从本地仓库安装当前源码版本：

```bash
./scripts/install-skills.sh --agent codex --source .
```

其他支持 `SKILL.md` 的 Agent 可以显式指定目录：

```bash
./scripts/install-skills.sh \
  --agent custom \
  --dest /absolute/path/to/agent/skills \
  --source .
```

不同 Agent 没有统一的 Skill 安装目录。custom 模式只负责复制正式 Skill 包，具体
发现时机和加载方式仍由目标 Agent 决定。

CLI 与 Skills 应固定在同一个 release tag。Skill 会通过
`obs-cli capabilities --output json --require ...` 协商实际能力，但把不断变化的
`main` Skills 与旧 CLI 混用仍可能导致不必要的 capability 不兼容。

### 3. 审计与手动升级

默认审计不会联网：

```bash
obs-cli doctor --agent codex --output json
```

它会检查可执行文件、配置、已注册 Vault 路径、正式 Codex Skills、托管 metadata、
本地 Skill 修改以及 CLI/Skill 版本是否一致。查询在线 Release 必须显式执行：

```bash
obs-cli doctor --agent codex --online --output json
obs-cli update check --output json
```

预演并手动执行经过校验的 CLI 升级：

```bash
obs-cli update apply --version v1.0.0 --dry-run --output json
obs-cli update apply --version v1.0.0 --output json
```

apply 会下载对应平台资产和 `checksums.txt`，验证候选版本与 `obs-cli/v1`
capabilities，将旧二进制保留为 `<path>.previous`，然后替换当前文件。该命令永远
不会自动运行。Windows 无法由运行中的进程安全替换自身，应继续使用
`scripts/install.sh --version <tag> --force`。

Skills 使用独立的显式升级：

```bash
bash /tmp/obs-cli-install-skills.sh \
  --agent codex \
  --version v1.0.0 \
  --upgrade \
  --dry-run

bash /tmp/obs-cli-install-skills.sh \
  --agent codex \
  --version v1.0.0 \
  --upgrade
```

首次安装会记录 digest 和托管版本。缺少 metadata 或 `SKILL.md` 已被本地修改时，
升级会停止。升级成功后，旧的托管 Skills 会保留在命令输出的备份目录中。

### 4. 注册 Vault

安装过程不会自动扫描或注册个人 Vault。应先只读发现和检查候选：

```bash
obs-cli vault discover --output json
obs-cli vault list --output json
```

先预演注册表变化，再对相同目标执行：

```bash
obs-cli vault add /absolute/path/to/vault \
  --name Personal \
  --set-default \
  --dry-run \
  --output json

obs-cli vault add /absolute/path/to/vault \
  --name Personal \
  --set-default \
  --output json
```

## 当前状态

首个 Agent-first 版本使用以下稳定顶层命令树：

```text
capabilities  vault  note  search  metadata
link          daily  doctor  update  template  batch
```

Agent 必须通过 `capabilities` 发现已实现 operation。`doctor` 默认只做离线审计，
`update` 只有在用户显式调用时才检查或修改版本。预留命名空间返回
`CAPABILITY_UNSUPPORTED`，不会回退到 picker、GUI 或 TTY 行为。

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

- [当前文档索引](./docs/README.md)
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
