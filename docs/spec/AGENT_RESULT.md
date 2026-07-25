# Agent Result v1

`miniobsidian.agent-result/v1` 是 Agent 集成层返回 Neovim 的结果契约。它描述已经
发生的外部结果和恢复信息，不是新的写入授权。

## 必需字段

- `request_id`：必须关联原 handoff；插件默认拒绝与最近请求不一致的结果。
- `status`：`success | partial | failed | cancelled | conflict`。
- `summary`：不包含正文的简要结果。
- `changes[]`：每项包含 Vault 相对 Markdown 路径、`revision_before`、
  `revision_after` 和摘要。
- `errors[]`：稳定错误码、消息、可选目标路径和人工恢复步骤。

`before_content` 是可选的 Agent 修改前正文。若对应笔记在 Neovim 中有 dirty
buffer，三方比较必须有该字段；缺失时插件保留本地和磁盘版本、阻止陈旧写入并
返回 `AGENT_BASE_MISSING`，不得伪造 base。

## 客户端处理规则

- `cancelled` 不打开 diff、不继续任何动作。
- `partial` / `failed` 必须显示 recovery checklist。
- 多文件结果先展示 changed-files picker。
- clean buffer 展示内存旧版本与磁盘 Agent 版本的 unified diff，不自动重载。
- dirty buffer 展示 base / Agent disk / Neovim local 三个版本。
- 采用磁盘版本前先保留本地 scratch snapshot；手动合并只创建可编辑 scratch
  buffer，不直接写回 Vault。

规范 Schema：
[agent-result-v1.schema.json](./schema/agent-result-v1.schema.json)。
