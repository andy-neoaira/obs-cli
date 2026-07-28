# P5：审计修复与 V2 收口

## 阶段目标

修复代码审计中已独立确认的正确性和持久性问题，删除未注册的 V1 命令，
并补强 capability 与 Skill 的自动一致性门禁。

审计来源：[review-audit.md](../review-audit.md)。本阶段只纳入复核后成立的问题，
不把候选意见直接当作实现要求。

## 进入条件

- P1、P2、P3 已完成。
- `make release-check` 在修复前通过，作为回归基线。
- 工作区中的用户改动已识别并保留。

P5 不依赖 P4 移动端人工验证，可以与 P4-T05/P4-T06 并行推进。

## 任务进度

阶段进度：`6 / 7`

- [x] [P5-T01 配置原子替换与目录持久化](./P5-T01-config-atomic-durability.md)
- [x] [P5-T02 Note List 隐藏文件隔离](./P5-T02-note-list-hidden-files.md)
- [x] [P5-T03 Obsidian 配置严格读取与错误可见性](./P5-T03-obsidian-config-errors.md)
- [x] [P5-T04 删除未注册的 V1 命令死代码](./P5-T04-remove-v1-command-dead-code.md)
- [x] [P5-T05 Capability 文档与实现一致性门禁](./P5-T05-capability-doc-gate.md)
- [x] [P5-T06 Skill 身份三方一致性门禁](./P5-T06-skill-identity-gate.md)
- [ ] [P5-T07 全量回归与审计关闭报告](./P5-T07-regression-audit-closure.md)

推荐顺序：

```text
P5-T01 ─┐
P5-T02 ─┼──────────────┐
P5-T03 ─┤              │
P5-T04 ─┤              ├─> P5-T07
P5-T05 ─┤              │
P5-T06 ─┘              │
```

T01、T02、T03 涉及安全与正确性，应优先于代码清理和文档补强。

## 明确不纳入

以下审计建议经复核后不作为 P5 修复项：

- 不改变 `note.move --dry-run` 的顶层响应结构。
- 不为 Vault registry dry-run 强行增加语义含糊的 `vault` 包装。
- 不在只读 Vault discovery 阶段强制使用内容路径 Resolver。
- 不把已在创建前后执行路径身份校验的 `os.MkdirAll` 判定为越界漏洞。
- 不在本阶段创建或推送正式 tag/release；正式发布仍受 P4 移动端门禁约束。

如未来需要调整这些设计，必须新增协议演进任务，不能顺手混入 P5。

## 阶段完成标准

- [ ] 配置文件 rename 后同步父目录，且保留现有配置锁与校验。
- [ ] 隐藏 Markdown 文件不会出现在列表中，也不会导致整个列表失败。
- [ ] Obsidian 配置缺失与损坏被区分，V2 写操作不会静默采用错误默认值。
- [ ] 未注册 V1 命令及其专属初始化副作用已删除。
- [ ] capability 实现与规范文档自动校验一致。
- [ ] Skill 目录名、frontmatter name 和 eval manifest 自动校验一致。
- [ ] `make release-check`、三入口自动 E2E 与相关负向测试全部通过。
- [ ] 审计关闭报告逐项记录修复提交、证据和保留决策。

## 通用执行规则

1. 开始任务前运行任务文档中的基线命令。
2. 先增加能复现问题的失败测试，再修改实现。
3. 每个任务只处理其“范围内”文件；发现新问题先记录，不顺手扩展。
4. 任务验证通过后更新子任务状态、此 README 进度及完成记录。
5. 一个任务一个独立提交；T07 只做回归、证据和状态收口。
6. 若修改协议或公共 Schema，必须先说明兼容性影响并更新相关文档。

## 任务模板

新增 P5 修复项必须基于 [TASK_TEMPLATE.md](./TASK_TEMPLATE.md)，至少填写：

- 可复现现象与证据
- 范围和非目标
- 分步执行及中间检查点
- 正向、负向和回归测试
- 可判断的验收标准
- 回滚方法与完成记录
