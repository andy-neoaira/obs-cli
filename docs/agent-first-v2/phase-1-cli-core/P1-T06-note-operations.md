# P1-T06：Note 原子操作 API

- 状态：`未开始`
- 负责人：`待分配`
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

- [ ] Agent 可完成读—分析—条件更新闭环。
- [ ] 默认不存在静默覆盖。
- [ ] patch 上下文不唯一或不匹配时不修改文件。
- [ ] 多行内容可从 stdin 或文件安全输入。
- [ ] 每个修改命令都有 dry-run。

## 验证

```bash
go test ./... -run 'Note|Append|Patch|Replace|Delete'
```

