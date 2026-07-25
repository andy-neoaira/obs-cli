# Agent Handoff v1

`miniobsidian.agent-handoff/v1` 是 Neovim 客户端交给 Agent 集成层的有界请求，
不是 CLI operation，也不代表用户已经授权执行写入。

## 边界

- payload 只携带稳定 Vault ID，不携带本机 Vault 绝对路径。
- `source.path`、`permissions.read_paths` 和 `permissions.write_paths` 都是规范化的
  Vault 相对 Markdown 路径。
- 默认 `allow_vault_scan=false`；Agent 不得因为收到 handoff 而扫描整个 Vault。
- `analyze` 的 `write_paths` 必须为空。
- `update` 只授权目标笔记路径；Agent 仍必须按指定 Skill 执行
  read → dry-run → authorize → apply → verify，handoff 本身不替代写入确认。
- `source.revision` 来自 handoff 前的 `note.get`，用于证据追踪和乐观并发控制。
- 未保存 buffer 只能用于 `analyze`，并明确标记 `buffer_modified=true`；
  `update` 必须在 buffer 保存后重新生成 handoff。

## 内容范围

`context.scope` 有三种：

- `none`：只交付路径、revision 和意图，Agent 按 `read_paths` 自行读取。
- `selection`：交付显式选区的内存文本及行号。
- `buffer`：仅 dirty buffer 的只读分析可交付完整内存快照。

任何包含内存文本的 handoff 应先向用户预览确认；大选区即使关闭普通内容确认，
也必须达到客户端配置阈值时再次确认。日志和 request ID 不得包含正文。

## Agent 接口

miniobsidian 通过用户配置的 `agent.handler(payload)` 交付 payload，避免强依赖任一
Agent 框架。handler 成功接收后，插件保存 `last_request` 并触发
`User MiniObsidianAgentHandoff`，事件只携带 request ID、mode 和相对路径。

规范 Schema：
[agent-handoff-v1.schema.json](./schema/agent-handoff-v1.schema.json)。
