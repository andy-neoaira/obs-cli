---
name: obsidian-vault-setup
description: 配置或检查 obs-cli V2 Vault 注册表。用户要求发现、列出、注册 Vault 或修改默认 Vault 时使用；不修改 Obsidian 官方配置。
---

# Obsidian Vault Setup

## 触发条件

- 用户要求发现或列出 Obsidian Vault。
- 用户明确要求把目录注册到 obs-cli。
- 用户要求查看或修改 obs-cli 的默认 Vault。

## 非触发条件

- 不用于创建、读取或整理笔记。
- 不用于修改 `.obsidian/` 或 Obsidian 桌面/移动端设置。
- 不处理打开方式等 V2 未声明的配置；明确说明当前 capability 不支持。

## 输入

- 只读查询无需额外输入。
- 注册需要绝对 `vault_path`，可选明确的 `name` 和 `set_default`。
- 修改默认项需要已注册 Vault 的 `id_or_name`。
- 路径、名称或目标不明确时先询问，不根据相似名称猜测。

## Capability 前置检查

每次先检查本场景所需 operation：

```bash
obs-cli capabilities --output json \
  --require vault.discover --require vault.list
obs-cli capabilities --output json --require vault.add
obs-cli capabilities --output json --require vault.set-default
```

只要求当前请求实际使用的 capability。缺失时停止并返回 CLI 版本、缺失 operation，以及“升级到提供 obs-cli/v2 对应 operation 的版本”的提示；不得回退到 `list-vaults`、`add-vault` 或 `set-default-vault`。

## 读取范围

- `vault discover` 只读 Obsidian 已有配置，用于发现候选项，不写入官方配置。
- `vault list` 只读 obs-cli V2 注册表。
- 不读取笔记正文。

```bash
obs-cli vault discover --output json
obs-cli vault list --output json
```

## 写入范围

- `vault add` 仅写 obs-cli V2 注册表。
- `vault set-default` 仅修改注册表中的默认 Vault。
- 禁止直接写 `.obsidian/`、删除 Vault、迁移配置或修改 Vault 内文件。

## 执行生命周期

严格执行 `discover → read → plan → authorize → apply → verify`：

1. `discover`：检查 capability，运行 `vault discover` 获取只读候选。
2. `read`：运行 `vault list`，记录当前注册项和默认项。
3. `plan`：生成稳定 request ID，对实际命令加 `--dry-run`，展示配置前后变化。
4. `authorize`：注册目标必须明确；修改默认 Vault 必须向用户展示变化并取得确认。
5. `apply`：使用相同路径、名称、选项和 request ID 执行一次。
6. `verify`：重新运行 `vault list`，验证规范化路径、ID 和默认项。

```bash
obs-cli vault add "<absolute-path>" --name "<name>" --dry-run \
  --request-id "<stable-id>" --output json
obs-cli vault add "<absolute-path>" --name "<name>" \
  --request-id "<stable-id>" --output json

obs-cli vault set-default "<id-or-name>" --dry-run \
  --request-id "<stable-id>" --output json
obs-cli vault set-default "<id-or-name>" \
  --request-id "<stable-id>" --output json
```

若用户明确要求“注册并设为默认”，可在 add 的计划和执行中使用 `--set-default`，仍需展示默认项变化。

## 授权策略

- 发现、列出和 dry-run 为只读，可直接执行。
- 用户明确给出路径并要求注册，可在展示 dry-run 后执行注册。
- 任何默认 Vault 变化都先展示当前值与目标值并确认。
- dry-run 出现 `/tmp` 到 `/private/tmp` 等系统 canonicalization 时展示映射并按同一文件身份验证；只有语义目标发生变化时停止。

## 冲突、重试与幂等

- 先 list：相同规范化路径已注册且名称、默认状态符合请求时返回 `no_change`。
- 同路径已注册但名称不同且没有 rename capability 时，不重复 add；报告无法满足的名称差异并请求决策。
- 请求结果不明时先重新 list；已生效则不得重复 add/set-default。
- Vault 配置 operation 当前不支持 `--if-match`；不得伪造 revision 或绕过 dry-run。若返回 `REVISION_CONFLICT`，按冲突停止处理，不得无条件重试。
- 部分完成（已注册但设默认失败）时返回 `partial`，保留实际状态并请求下一步。
- 所有值通过结构化 argv 传递，禁止插值生成 shell 命令；长文本仅允许通过 stdin 或文件，尽管本 Skill 不需要正文。

## 结果摘要

```yaml
status: success | no_change | conflict | partial | failed
vault:
  id: <resolved-id>
  name: <resolved-name>
  path: <canonical-path>
operation: discover | list | add | set-default
planned: [<dry-run changes>]
applied: [<registry changes>]
verified:
  registered: true | false
  default: true | false
conflicts: []
warnings: []
next_actions: []
```
