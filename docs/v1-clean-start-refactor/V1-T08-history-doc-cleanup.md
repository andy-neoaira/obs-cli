# V1-T08：删除历史 V2 文档并收口当前文档

- 状态：`实现完成，待提交`
- 优先级：`中`
- 涉及项目：`obs-cli`、`miniobsidian.nvim`
- 依赖：V1-T02～V1-T07

## 目标

让当前文档只描述 V1 产品，不再把 Agent-first 方向表述为旧产品的 V2 升级；历史任务
记录不再充当当前事实源。

## 文档分类

### 当前事实源

必须改写为 V1：

- `README.md`
- `README_CN.md`
- `docs/COMMAND_REFERENCE.md`
- `docs/spec/`
- `docs/JOINT_USAGE.md`
- `docs/TROUBLESHOOTING_AND_RECOVERY.md`
- `docs/compatibility.json`
- Skills 和 eval 文档
- `miniobsidian.nvim` 当前 README/help

### 历史执行记录

当前包括：

- `docs/agent-first-v2/`
- `docs/releases/V2_RC_VALIDATION.md`
- `docs/releases/AGENT_FIRST_V2_ACCEPTANCE.md`
- `docs/migration/V1_TO_V2.md`
- 引用上述目录的 ADR 和 release 文档

## 强制决策

删除旧执行计划、审计收口记录和 V1→V2 迁移叙事，不在仓库内建立历史归档副本。需要
追溯时使用 Git 历史。许可证、派生关系和第三方通知不属于删除范围。

## 禁止事项

- 不对历史文档做无意义的全局 V2→V1 改写，从而伪造历史。
- 不保留“V1 到 V2”作为当前用户迁移主线。
- 不删除许可证、派生说明和第三方通知。
- 不把旧文档继续链接为当前规范。
- 不扩大归档 allowlist 以掩盖当前源码残留。

## 执行步骤

1. 建立当前事实源和历史记录文件清单。
2. 先更新当前 README、命令参考、规范和恢复文档。
3. 将产品描述统一为“obs-cli V1 / Agent-first Obsidian CLI”。
4. 删除 V1→V2 迁移文档，不提供新旧产品迁移承诺。
5. 删除 `docs/agent-first-v2` 及不再适用的旧 release/migration 文档。
6. 重写 ADR 中当前架构名称，历史动机留在 Git history。
7. 更新所有相对链接。
8. 检查中英文文档对称性。
9. 运行链接和 Schema 引用检查。
10. 检索当前文档中的禁止名称。

## 验收标准

- [ ] 当前 README 和规范只描述 V1。
- [ ] 当前文档没有 V1→V2 产品叙事。
- [ ] 旧阶段计划已删除，仓库中没有复制的历史归档。
- [ ] 当前导航不链接已删除的旧规范。
- [ ] 所有相对链接有效。
- [ ] LICENSE、派生说明和 THIRD_PARTY_NOTICES 保留。
- [ ] 中英文文档关键版本、协议和配置路径一致。

## 验证命令

```bash
rg -n -i 'Agent-first V2|obs-cli/v2|config-v2\.json|v2\.0\.0-rc\.1|V1_TO_V2' \
  README.md README_CN.md docs/spec docs/COMMAND_REFERENCE.md \
  docs/JOINT_USAGE.md docs/TROUBLESHOOTING_AND_RECOVERY.md \
  skills /Users/andy/github/miniobsidian.nvim/README.md \
  /Users/andy/github/miniobsidian.nvim/README.en.md \
  /Users/andy/github/miniobsidian.nvim/doc

make schema-check
make skill-lint
git diff --check
```

当前文档检索必须无输出。

## 回滚

文档删除使用单独提交，以便通过 Git 恢复审计记录。即使回滚删除提交，也不得恢复
README 对旧规范的当前链接。

## 完成记录

- 完成日期：`2026-07-28（工作区实现）`
- obs-cli 提交：`未提交`
- miniobsidian.nvim 提交：`未提交`
- 删除文件：`待填写`
- 保留的合规文档：`待填写`
