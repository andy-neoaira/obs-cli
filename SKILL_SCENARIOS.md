# obs-cli 场景 Skill 规划清单

> 用途：把 obs-cli 的常用组合使用方式抽象为 Claude Code skills，供后续筛选、合并、改名或落地实现。
>
> 当前文件只做规划，不包含具体 skill 实现代码。

---

## 推荐优先级总览

### 第一批：最高频、最基础

1. `obsidian-vault-setup`
2. `obsidian-capture`
3. `obsidian-daily-log`
4. `obsidian-knowledge-search`

覆盖闭环：配置 Vault → 写入内容 → 每日记录 → 检索上下文。

### 第二批：工作流增强

5. `obsidian-meeting-note`
6. `obsidian-project-note`
7. `obsidian-frontmatter-workflow`

覆盖结构化归档、项目知识库、metadata 状态流。

### 第三批：整理 / 审计 / 自动化

8. `obsidian-inbox-triage`
9. `obsidian-vault-audit`
10. `obsidian-cli-automation`

覆盖维护、盘点、脚本化集成。

---

## 1. obsidian-vault-setup

**定位**：初始化 / 配置 Vault。

**触发语义**：

- “帮我配置 obs-cli”
- “把这个目录注册为 Obsidian vault”
- “设置默认 vault”
- “设置默认用 editor 打开”

**覆盖命令**：

```bash
obs-cli add-vault <path> [--set-default]
obs-cli set-default-vault [name] [--open-type obsidian|editor]
obs-cli list-vaults [--default] [--path-only|--json]
```

**典型流程**：

1. 查看当前已注册 vault：`obs-cli list-vaults`
2. 如用户提供路径，则执行：`obs-cli add-vault <path>`
3. 根据需要设为默认：`--set-default` 或 `obs-cli set-default-vault <name>`
4. 根据需要设置默认打开方式：`obs-cli set-default-vault --open-type editor`
5. 输出当前默认 vault：`obs-cli list-vaults --default`

**Skill 输入建议**：

- `vault_path`
- `set_default: boolean`
- `open_type: obsidian|editor`

**输出建议**：

- 当前默认 vault 名称
- 当前默认 vault 路径
- 当前默认打开方式
- 后续可用命令示例

---

## 2. obsidian-capture

**定位**：快速捕获内容到 Inbox 或指定路径。

**触发语义**：

- “记到 Obsidian”
- “保存到我的笔记”
- “放到 Inbox”
- “把这段内容写成一条笔记”

**覆盖命令**：

```bash
printf '%s' "<text>" | obs-cli create <note> --content-file -
printf '%s' "<text>" | obs-cli create <note> --content-file - --append
obs-cli create <note> --open --editor
obs-cli open <note> --editor
```

**典型流程 A：创建单独捕获笔记**：

```bash
printf '%s' "..." | obs-cli create "Inbox/<title>.md" --content-file -
obs-cli open "Inbox/<title>.md" --editor
```

**典型流程 B：追加到统一 Inbox 文件**：

```bash
printf '%s' "- captured item" | obs-cli create "Inbox.md" --content-file - --append
```

**适用场景**：

- 临时想法
- AI 输出结果落库
- 代码 review 结论存档
- Web research 摘要存档
- 待处理任务快速记录

**Skill 输入建议**：

- `title`
- `content`
- `target_path`，默认 `Inbox/<title>.md` 或 `Inbox.md`
- `mode: create|append`
- `open_after: boolean`

**输出建议**：

- 创建 / 追加的 note path
- 是否已打开编辑器
- 后续整理建议

---

## 3. obsidian-daily-log

**定位**：每日笔记记录。

**触发语义**：

- “写到今天的 daily note”
- “记录今天完成了什么”
- “把这次 session 总结追加到日记”
- “写一条工作日志”

**覆盖命令**：

```bash
obs-cli daily
printf '%s' "<text>" | obs-cli daily --content-file -
printf '%s' "<text>" | obs-cli daily --content-file - --editor
obs-cli daily --editor
```

**典型流程**：

```bash
printf '%s' "- 09:30 standup" | obs-cli daily --content-file -
printf '%s' "## Session Summary\n..." | obs-cli daily --content-file -
obs-cli daily --editor
```

**适用场景**：

- 工作日志
- 站会记录
- Claude Code session 总结
- 每日计划 / 回顾
- 临时流水账记录

**Skill 输入建议**：

- `content`
- `section`，例如 `Work Log` / `Meeting` / `Summary` / `TODO`
- `format: bullet|section|raw`
- `open_after: boolean`

**输出建议**：

- 已追加内容摘要
- daily note 所在路径，如果可推断
- 建议的 Markdown 格式

---

## 4. obsidian-knowledge-search

**定位**：从 Vault 检索知识并返回上下文。

**触发语义**：

- “在我的 Obsidian 里找”
- “搜索 vault”
- “找相关笔记”
- “把相关笔记内容拿出来给我”
- “从我的知识库里查一下”

**覆盖命令**：

```bash
obs-cli search-content <query> --format json --page 1 --page-size <n>
obs-cli print <note>
obs-cli print <note> --mentions
obs-cli open <note> [--editor]
```

**典型流程**：

1. 搜索关键词：
   ```bash
   obs-cli search-content "keyword" --format json --page 1 --page-size 10
   ```
2. 根据 JSON 结果选择相关笔记。
3. 打印笔记内容：
   ```bash
   obs-cli print "matched-note.md"
   ```
4. 必要时附带反向链接：
   ```bash
   obs-cli print "matched-note.md" --mentions
   ```

**适用场景**：

- 为 AI 回答提供本地知识库上下文
- 查找过去的决策、会议记录、debug 记录
- RAG-lite 场景
- 主题研究前的本地资料检索

**Skill 输入建议**：

- `query`
- `top_n`
- `vault`
- `include_mentions: boolean`
- `open_selected: boolean`

**输出建议**：

- 命中的 note 列表
- 相关内容摘要
- 已打印 / 建议读取的 note
- 可选 mentions 摘要

## 6. obsidian-project-note

**定位**：项目级笔记创建和维护。

**触发语义**：

- “给这个项目建一份 Obsidian 笔记”
- “把代码分析结果写入 Vault”
- “更新项目知识库”
- “把这次项目 review 存档”

**覆盖命令**：

```bash
printf '%s' "<markdown>" | obs-cli create <project-note> --content-file - [--overwrite|--append]
obs-cli frontmatter <project-note> --edit --key project --value <name>
obs-cli frontmatter <project-note> --edit --key status --value active
obs-cli search-content <project-name> --format json
obs-cli open <project-note> --editor
```

**典型流程**：

```bash
printf '%s' "..." | obs-cli create "Projects/obs-cli/Overview.md" --content-file - --overwrite
obs-cli frontmatter "Projects/obs-cli/Overview.md" --edit --key type --value project
obs-cli frontmatter "Projects/obs-cli/Overview.md" --edit --key project --value obs-cli
obs-cli frontmatter "Projects/obs-cli/Overview.md" --edit --key status --value active
```

**适用场景**：

- 代码项目分析后写入 Obsidian
- 技术文档同步到 Vault
- 项目状态追踪
- 架构、命令、风险点归档

**Skill 输入建议**：

- `project_name`
- `note_type: overview|architecture|session-log|todo|decision`
- `content`
- `mode: overwrite|append|create-only`
- `target_folder`，默认 `Projects/<project_name>`

**输出建议**：

- project note path
- 更新模式
- frontmatter 字段
- 后续建议维护的关联笔记

---

## 8. obsidian-inbox-triage

**定位**：整理 Inbox 笔记。

**触发语义**：

- “帮我整理 Inbox”
- “把这个临时笔记移动到项目目录”
- “归档这条笔记”
- “把 Inbox 里的内容分类”

**覆盖命令**：

```bash
obs-cli list Inbox
obs-cli move <source-note> <target-note>
obs-cli frontmatter <target-note> --edit --key status --value organized
obs-cli open <target-note> --editor
```

**典型流程**：

```bash
obs-cli list "Inbox"
obs-cli move "Inbox/tmp.md" "Projects/MyProject/final.md"
obs-cli frontmatter "Projects/MyProject/final.md" --edit --key status --value organized
obs-cli open "Projects/MyProject/final.md" --editor
```

**适用场景**：

- 每日 / 每周 Inbox 清理
- 临时记录转为正式知识卡片
- 将捕获内容归档到 Projects / Areas / Notes
- 重命名但保持链接不坏

**Skill 输入建议**：

- `source_note`
- `target_path`
- `classification`
- `status`
- `open_after: boolean`

**输出建议**：

- 移动前后路径
- 更新的状态字段
- 是否需要补充 frontmatter / tags

---

## 9. obsidian-vault-audit

**定位**：Vault 内容盘点 / 巡检。

**触发语义**：

- “检查我的 vault”
- “找所有 TODO”
- “看看哪些笔记提到了某个主题”
- “帮我盘点项目笔记”

**覆盖命令**：

```bash
obs-cli list [path]
obs-cli search-content <query> --format json --page 1 --page-size 100
obs-cli print <note> --mentions
```

**典型流程**：

```bash
obs-cli list
obs-cli list "Projects"
obs-cli search-content "TODO" --format json --page 1 --page-size 100
obs-cli print "note.md" --mentions
```

**适用场景**：

- 定期知识库检查
- 查找 TODO / FIXME / draft / stale note
- 查找某主题被哪些笔记引用
- 生成整理建议

**Skill 输入建议**：

- `directory`
- `search_terms`
- `audit_type: todo|topic|mentions|status|structure`
- `max_results`

**输出建议**：

- 命中列表
- 分类统计
- 高优先级整理建议
- 可执行的后续 obs-cli 命令
