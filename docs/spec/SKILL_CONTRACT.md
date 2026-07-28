# obs-cli V1 Skill 契约

## 1. 适用范围

本契约适用于所有通过 `obs-cli` 读取或修改 Obsidian Vault 的场景化 Skill。Skill 是 Agent 工作流，不是命令清单；它必须限制读取和写入范围，并对执行结果负责。

所有 Skill 必须以 [`skills/_template/SKILL.md`](../../skills/_template/SKILL.md) 为起点。
不符合当前 Capability 契约的 Skill 会被 lint 明确报告，发布门禁使用严格模式禁止遗留项。

## 2. 文件与元数据

- 每个 Skill 使用小写连字符目录名，入口为 `SKILL.md`。
- YAML frontmatter 只包含 `name` 和 `description`。
- `name` 与目录名一致；`description` 同时说明能力和触发上下文。
- 正文必须包含模板中的十个二级章节，章节名不得自行改写。

## 3. 生命周期

```text
discover → read → plan → authorize → apply → verify
```

| 阶段 | 必须完成的工作 |
|---|---|
| discover | 解析 Vault 和输入；用 `capabilities --output json --require` 检查所需 operation |
| read | 限定读取范围；修改前取得目标内容和 revision |
| plan | 修改型操作执行 dry-run，记录目标、diff、风险、前置条件；只读操作声明无写入 |
| authorize | 检查用户授权是否覆盖实际目标、范围和风险 |
| apply | 使用计划中的目标和 revision 执行一次写入 |
| verify | 重新读取目标，验证 revision、内容和场景不变量 |

任何阶段失败都停止后续阶段。不得因为旧命令“看起来可用”而绕过 capability 检查。

## 4. 授权边界

以下条件全部满足时，用户原始请求可视为低风险写入授权：

1. 用户明确要求修改；
2. Vault、目标和写入方式无歧义；
3. dry-run 未扩大范围；
4. 操作不涉及覆盖、删除、批量移动或默认 Vault 变更。

否则必须先展示计划并请求确认。只读分析、capability 检查和 dry-run 不需要额外确认。Skill 不得把“分析”“比较”“建议”解释为写入授权。

## 5. 并发、重试与幂等

- 修改既有笔记和删除操作必须携带读取阶段得到的 `--if-match` revision。
- `REVISION_CONFLICT` 是正常并发结果：停止写入、重新读取、展示差异并请求用户决策。
- 禁止去掉 `--if-match` 后重试，禁止自动覆盖其他入口的修改。
- 使用稳定目标与 request ID；网络或进程错误后先验证是否已生效，再决定是否重试。
- 批量操作逐项记录结果。发生部分失败时停止依赖失败项的后续步骤，并返回 `partial`。

## 6. Shell 与内容安全

- 用户原文、多行 Markdown 和长文本只通过 stdin 或受控临时文件传入。
- 不把内容插入 shell 命令字符串，不使用命令替换解释用户内容。
- 路径作为独立参数传递，并遵守 Vault 相对路径约束。
- 不记录正文、密钥或不必要的个人笔记内容到日志。

## 7. 结果与验证

Skill 必须返回 `status`、解析后的 Vault、读取项、计划、实际修改、验证、冲突、警告和后续动作。只读 Skill 的 `applied` 必须为空；修改型 Skill 的成功结果必须包含写后读取证据。

允许的状态为：

- `success`：计划内操作已完成且验证通过；
- `no_change`：无需修改；
- `conflict`：revision 冲突，未覆盖；
- `partial`：部分完成，逐项列明；
- `failed`：未达到场景目标。

命令退出成功不等于场景成功；verify 阶段失败时不得返回 `success`。

## 8. Lint 门禁

```bash
./scripts/lint-skills.sh
./scripts/lint-skills.sh --strict
./scripts/lint-skills.sh path/to/SKILL.md
```

- 默认模式严格检查模板和已经声明 Capability 章节的 Skill，并将旧格式 Skill 报告为警告。
- 指定文件时始终严格检查，适合开发单个 Skill。
- `--strict` 将所有旧格式 Skill 视为错误，供最终发布门禁使用。
