# 从 V1 迁移到 Agent-first V2

## 1. 升级性质

V2 是破坏性升级。V1 命令实现只保留在 Git 历史中用于审计，当前源码不再包含
未注册的 V1 Cobra 命令。V2 不提供未文档化兼容别名，也不会根据 TTY、cwd 或本机
GUI 状态改变操作语义。

升级前建议保留当前 V1 二进制，并为重要 Vault 创建同步服务版本点或文件系统快照。

## 2. 命令映射

| V1 | V2 | 迁移说明 |
|---|---|---|
| `add-vault <path>` | `vault add <path>` | 使用 V2 独立注册表；先 dry-run |
| `list-vaults` | `vault list` / `vault discover` | 已注册 Vault 与只读 Obsidian 发现分离 |
| `remove-vault <name>` | `vault remove <id-or-name>` | 不删除 Vault 文件 |
| `set-default-vault` | `vault set-default <id-or-name>` | 不再交互选择 |
| `create <note>` | `note create <path> --content-file <file|->` | 默认 must-not-exist |
| `print <note>` | `note get <path>` | JSON 同时返回 content、Frontmatter、revision |
| `list [path]` | `note list` | 返回稳定逻辑路径数组 |
| `delete <note>` | `note delete <path> --if-match <revision>` | 需要条件删除并保留私有恢复副本 |
| `move <old> <new>` | `note move <old> <new> --if-match <revision>` | 链接重写纳入事务 |
| `frontmatter` | `note get` + `note patch/replace` | 使用 revision 与上下文条件 |
| `search` | `search content <query>` | 有界分页并返回 path/revision/line/snippet 证据 |
| `search-content` | `search content <query>` | 使用统一 V2 JSON envelope |
| `daily` | `daily get/create/append` | 遵循 Obsidian folder/format/template |
| `open` | 无 V2 替代 | GUI/编辑器启动移出 Agent 核心 |

V1 的 `--editor`、模糊 picker、Obsidian URI 启动和自然语言 stdout 没有 V2 兼容模式。

## 3. 配置迁移

V2 不修改 Obsidian 官方配置。执行：

```bash
obs-cli vault discover --output json
obs-cli vault migrate --dry-run
obs-cli vault migrate
obs-cli vault list
```

迁移只允许在 V2 注册表尚无 Vault 时执行。V1 名称无法唯一解析时会返回 migration warning；Agent 必须显式修正，不得猜测。

## 4. Skill 迁移规则

1. 首次执行前调用 `capabilities --require <operation>`。
2. 显式传入 `--vault` 或配置明确默认 Vault。
3. 多行内容只通过 `--content-file -`/文件传递。
4. 写入前先 `note get` 并保存 revision。
5. 先 dry-run，再使用相同业务输入 apply。
6. `REVISION_CONFLICT` 后重新读取和分析；禁止去掉 `--if-match` 重试。
7. 不解析 message、帮助文本或 stderr；只读取 JSON envelope、operation、error code 和 details。

## 5. 已知限制

- CLI 不负责模板交互 UI、批量自由编辑或 GUI doctor；模板渲染由 Daily operation
  和独立客户端按共同规范处理。
- 外部 Obsidian、Neovim 和同步服务不遵守 CLI 协作锁；提交前 revision 复核会缩小但不能消除最终原子替换前的极短竞态窗口。
- 网络文件系统和云盘虚拟文件系统可能不提供本地 rename/flush 的同等持久性。
- move 重写支持 V2 规范中的 Wikilink 与普通 Markdown 文件链接；Vault 绝对风格链接和外部 scheme 不重写。
- V2 不启动 Obsidian 或编辑器。客户端 UI 集成应由 Neovim 插件或其他独立客户端负责。

## 6. 回滚

如果 V2 RC 不满足工作流：

1. 停止 Agent/Skill 的 V2 写入。
2. 保存失败响应中的 transaction ID 和恢复步骤。
3. 对 `PARTIAL_FAILURE` 先完成恢复，不要直接重放操作。
4. 将 PATH 中的 `obs-cli` 切回升级前保存的 V1 二进制。
5. V2 注册表与 Obsidian 官方配置独立；不要删除或覆盖官方 `obsidian.json`。
6. 如需重新测试 V2，使用临时 Vault 和独立二进制名。

回滚二进制不会自动撤销已成功提交的 Markdown 修改；应使用返回 revision、同步历史或升级前快照恢复业务内容。
