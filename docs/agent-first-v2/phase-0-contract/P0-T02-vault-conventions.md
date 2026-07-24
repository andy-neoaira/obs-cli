# P0-T02：Vault 共同约定

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli`、`miniobsidian.nvim`
- 依赖：P0-T01

## 目标

定义两个实现共同遵守的 Vault 行为规范，并将规范标记为 `vault-contract/v1`。

## 实施步骤

1. 新建 `docs/spec/VAULT_CONVENTIONS.md`。
2. 定义 Vault 根目录、相对路径、规范化、大小写和符号链接规则。
3. 定义忽略目录、Markdown 扩展名、编码和换行约定。
4. 定义 note ID、文件路径、标题和 Frontmatter 的关系。
5. 定义 Wikilink 的 path、alias、heading、block 与同名消歧。
6. 定义附件目录与相对链接生成规则。
7. 定义 Daily Note 的目录、日期格式、模板和已存在文件行为。
8. 定义移动、删除、链接重写和外部修改规则。
9. 在两个仓库文档中声明支持的规范版本。

## 交付物

- `obs-cli/docs/spec/VAULT_CONVENTIONS.md`
- 两个项目的规范版本声明

## 验收标准

- [x] 每条规范具有唯一章节编号。
- [x] 明确 POSIX、macOS 大小写不敏感文件系统和 Windows 的差异。
- [x] 同名笔记不能被静默选中。
- [x] Daily Note 在两个入口产生相同目标路径和模板语义。
- [x] 路径逃逸与符号链接逃逸被明确禁止。
- [x] 破坏性规范变更必须升级主版本。

## 验证

```bash
rg -n "vault-contract/v1|Wikilink|Daily Note|符号链接|同名" \
  /Users/andy/github/obs-cli/docs/spec/VAULT_CONVENTIONS.md
```

## 验证记录

- 2026-07-24：共检查 57 个 `VC-*` 规则章节，重复规则 ID 为 0。
- 2026-07-24：路径、平台差异、同名消歧、Daily Note、符号链接与主版本规则检查通过。
- 2026-07-24：两个项目均声明 `target_contract` 与 `implemented_contract`，未将待实现规范虚报为已符合。
- 2026-07-24：本地 Markdown 链接和两个仓库 `git diff --check` 均通过。
