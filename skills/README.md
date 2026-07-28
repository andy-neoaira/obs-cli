# obs-cli Skills

本目录保存面向 agent 的 obs-cli 场景化 skill。每个子目录包含一个 `SKILL.md`，用于描述触发场景、输入字段、推荐命令和安全规则。

## 当前 Skill

- `obsidian-vault-setup`：发现、注册并设置默认 vault。
- `obsidian-capture`：把内容快速捕获到 Inbox 或指定笔记。
- `obsidian-daily-log`：把日志、总结、会议记录追加到 Daily Note。
- `obsidian-knowledge-search`：只读搜索 vault 并提取相关笔记上下文。
- `obsidian-project-note`：创建或维护项目级笔记和 frontmatter。
- `obsidian-inbox-triage`：整理 Inbox，移动笔记并更新状态 metadata。
- `obsidian-vault-audit`：只读盘点 vault 结构、TODO、status 和主题命中。

## 安全约定

- 写入长文本或多行 Markdown 时，优先使用 `--content-file -` 从 stdin 传入。
- 不要把用户原文直接拼接到 `--content "<text>"` 中，避免 shell 引号、换行和特殊字符破坏命令。
- destructive 操作只在用户明确要求时执行；当前 skill 默认不调用 `delete`。
- 对推断出来的移动目标、覆盖写入目标或默认 vault 变更，执行前应确认。

## 验收建议

发布前应至少用 `SKILL_SCENARIOS.md` 中的场景做人工或自动化验收：

- setup：发现、注册并设置默认 vault。
- capture：创建 Inbox 笔记、追加 Inbox、stdin 多行输入。
- daily：追加 daily、只打开 daily、模板配置存在时创建 daily。
- search/audit：JSON 搜索、分页、读取 top notes、mentions。
- triage/project：移动笔记、更新 frontmatter、避免覆盖已有笔记。
