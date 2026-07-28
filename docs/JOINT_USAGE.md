# obs-cli × miniobsidian.nvim 联合使用指南

## 产品边界

Markdown Vault 是唯一内容事实源。Obsidian、`obs-cli` 和
`miniobsidian.nvim` 是三个同级客户端：

- 插件不依赖 CLI 即可完成本地创建、跳转、搜索、链接、模板和 Daily Note。
- Agent 通过 CLI 与场景化 Skills 读取、比较和安全更新笔记。
- CLI Adapter 只提供移动、审计、Agent handoff/result 等高级协同，不改变插件
  的独立产品边界。

兼容关系以 [compatibility.json](./compatibility.json) 为机器事实源。当前发布物均
是候选状态，尚未创建正式 tag。

## 安装与启用

两个项目独立安装、独立升级。先确认插件本地模式正常，再选择是否启用 CLI：

```lua
require("miniobsidian").setup({
  cli = {
    enabled = "auto",
    command = "obs-cli",
    timeout_ms = 3000,
  },
})
```

`auto` 在 CLI 缺失或不兼容时保留纯插件模式；`true` 会在 health 中把问题提升为
警告；`false` 完全关闭 Adapter：

```lua
require("miniobsidian").setup({
  cli = { enabled = false, command = "obs-cli", timeout_ms = 3000 },
})
```

运行 `:checkhealth miniobsidian` 查看状态。Adapter 依次校验：

1. `capabilities --output json` 可执行且返回合法 envelope；
2. `obs-cli/v1`；
3. `vault-contract/v1`；
4. 当前操作所需 operation。

只有四层均满足时才启用对应高级能力。

## 三种运行状态

| 状态 | 插件本地功能 | CLI 高级功能 | 用户动作 |
|---|---|---|---|
| 无 CLI/已禁用 | 可用 | 禁用 | 可保持现状 |
| 兼容 CLI | 可用 | 按 operation 启用 | 正常使用 |
| 不兼容/异常 CLI | 可用 | 安全禁用 | 检查 health，升级、降级或关闭 Adapter |

## 升级与降级

1. 升级前备份 CLI 二进制、插件 lockfile/commit 和重要 Vault。
2. 在临时 Vault 运行 `make release-check`、插件 CI 和三入口 E2E。
3. CLI 与插件可以分开升级；升级后执行 `:ObsidianCLIRefresh`。
4. 不兼容时先设 `cli.enabled = false`，再切回上一个二进制或插件 commit。
5. 升降级不要求改写 Markdown、Frontmatter、Wikilink 或 Daily Note 格式。

V1 CLI 自有 registry 与 Obsidian 官方配置分离。降级二进制不会撤销已经成功提交的
笔记变更；业务内容恢复必须使用同步历史、版本控制或备份。

## Agent 使用约束

- Skill 先检查 capabilities，再执行读操作。
- 修改前读取 revision，先 dry-run，再携带 revision/plan hash apply。
- `REVISION_CONFLICT` 后重新读取并重新分析，禁止移除前置条件强行重试。
- dirty Neovim buffer 不交给 Agent 写入；Agent 更新后用 diff/三方视图处理。
- 移动端或同步服务未完成同步时暂停 Agent 写入，避免把延迟误判为删除。
