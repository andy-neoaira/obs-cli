# obs-cli Skills

本目录保存面向 agent 的 obs-cli 场景化 skill。每个子目录包含一个 `SKILL.md`，用于描述触发场景、输入字段、推荐命令和安全规则。

## 当前 Skill

- `obsidian-vault-setup`：发现、注册并设置默认 vault。
- `obsidian-capture`：把内容快速捕获到 Inbox 或指定笔记。
- `obsidian-compare-notes`：读取并比较多篇笔记，不执行写入。
- `obsidian-daily-log`：把日志、总结、会议记录追加到 Daily Note。
- `obsidian-inbox-triage`：整理 Inbox，安全移动笔记并更新 metadata。
- `obsidian-knowledge-search`：只读搜索 vault 并提取相关笔记上下文。
- `obsidian-knowledge-synthesis`：基于有界来源生成或更新综合笔记。
- `obsidian-project-note`：创建或维护项目级笔记和 frontmatter。
- `obsidian-project-status`：汇总项目证据并追加状态记录。
- `obsidian-safe-note-update`：使用 revision 和 patch 安全更新现有笔记。
- `obsidian-vault-audit`：只读盘点 vault 结构、TODO、status 和主题命中。

## 安全约定

- 写入长文本或多行 Markdown 时，优先使用 `--content-file -` 从 stdin 传入。
- 不要把用户原文直接拼接到 `--content "<text>"` 中，避免 shell 引号、换行和特殊字符破坏命令。
- destructive 操作只在用户明确要求时执行；当前 skill 默认不调用 `delete`。
- 对推断出来的移动目标、覆盖写入目标或默认 vault 变更，执行前应确认。

## 验收建议

当前确定性场景清单位于
[`evals/scenarios.json`](./evals/scenarios.json)，兼容要求位于
[`evals/compatibility-matrix.md`](./evals/compatibility-matrix.md)。发布前运行：

```bash
make skill-evals
```

模型主观质量评测与确定性安全门禁分开记录，不使用已经落地前的规划清单作为验收依据。
