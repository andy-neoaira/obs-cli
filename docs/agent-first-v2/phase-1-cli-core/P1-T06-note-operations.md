# P1-T06：Note 原子操作 API

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli`
- 依赖：P1-T02、P1-T03、P1-T04、P1-T05

## 目标

提供适合 Agent 组合的 note 原子操作，减少整篇覆盖。

## 实施步骤

1. 实现 `note list/get/create/append/patch/replace/delete`。
2. `get` 返回原始内容、Frontmatter、路径和 revision。
3. `create` 默认拒绝覆盖，重复请求可通过 request/idempotency 规则识别。
4. `append` 明确换行和目标章节规则。
5. `patch` 支持上下文 patch，并在上下文不匹配时失败。
6. `replace/delete` 要求 `--if-match` 或显式危险选项。
7. 所有操作接入路径、原子存储、JSON 和 dry-run。
8. 增加 stdin 输入，禁止 Skill 将长 Markdown 拼进 shell 参数。

## 交付物

- V2 note 命令
- 原子业务 API
- 正常、冲突、幂等和异常测试

## 验收标准

- [x] Agent 可完成读—分析—条件更新闭环。
- [x] 默认不存在静默覆盖。
- [x] patch 上下文不唯一或不匹配时不修改文件。
- [x] 多行内容可从 stdin 或文件安全输入。
- [x] 每个修改命令都有 dry-run。

## 验证

```bash
go test ./... -run 'Note|Append|Patch|Replace|Delete'
```

## 完成记录

- P5-T02 补齐 `note.list` 隐藏条目隔离：隐藏目录、文件和链接均不会出现在结果中，
  也不会让普通笔记列表整体失败；显式访问隐藏路径仍由路径策略拒绝。

- 新增 `note list/get/create/append/patch/replace/delete` V2 JSON 命令。
- 业务层统一接入 Vault path policy、稳定快照、revision 前置条件、原子写入和可恢复删除。
- `get` 同一快照返回原始正文、Frontmatter、逻辑路径与 revision。
- create 使用 must-not-exist，并在冲突 details 中返回现有/请求 revision 与 `same_content`，供 Agent 判断安全重试。
- append 固化边界换行与唯一 ATX section 规则，并忽略 fenced code block 内的伪标题。
- patch 要求 revision 和唯一原始字节上下文；不匹配、歧义、revision 冲突均保持文件不变。
- replace/delete 默认要求 `--if-match`，危险绕过必须显式使用 `--unsafe-no-if-match`。
- 所有正文输入只接受文件或 stdin；所有修改命令均有无写入副作用的 dry-run。
- 完整命令与 Agent 闭环见 `docs/spec/NOTE_OPERATIONS.md`。
