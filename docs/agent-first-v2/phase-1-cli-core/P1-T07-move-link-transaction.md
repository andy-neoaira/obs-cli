# P1-T07：Move 与链接重写事务

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`obs-cli`
- 依赖：P1-T03、P1-T06、P0-T05

## 目标

安全移动笔记并更新受影响链接，避免目标覆盖、错误重写和半完成状态。

## 实施步骤

1. 分离 plan 与 apply，计划中列出源、目标和所有待更新文件。
2. 目标存在时默认失败，不依赖平台特定 rename 覆盖行为。
3. 使用规范中的 Wikilink/Markdown link 解析器，不做无上下文全局字符串替换。
4. apply 前重新检查所有 revision。
5. 将原文件内容和权限保存到受控临时区。
6. 任一写入失败时执行回滚；无法回滚时返回 `PARTIAL_FAILURE` 和恢复清单。
7. 支持 `--dry-run` 输出精确 diff。
8. 增加同名、相对链接、alias、heading 和并发修改测试。

## 交付物

- move/rewrite plan
- 事务执行与回滚机制
- 多文件集成测试

## 验收标准

- [ ] 目标文件不会被静默覆盖。
- [ ] 计划后发生外部修改时 apply 被拒绝。
- [ ] 不相关文本和代码块不会被误改。
- [ ] 正常失败能完整恢复原状态。
- [ ] `PARTIAL_FAILURE` 包含机器可读恢复步骤。

## 验证

```bash
go test ./... -run 'Move|Rewrite|Rollback|PartialFailure'
```

