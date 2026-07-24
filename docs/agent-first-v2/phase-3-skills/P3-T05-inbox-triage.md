# P3-T05：迁移 Inbox Triage

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`obs-cli Skills`
- 依赖：P3-T01、P1-T07

## 目标

将 Inbox 整理升级为可审查、可回滚、支持冲突检测的多文件工作流。

## 实施步骤

1. 只读扫描 Inbox，记录每篇 revision。
2. 为分类、metadata、链接和移动生成统一 plan。
3. 展示每篇笔记的目标路径和修改摘要。
4. 用户授权后按事务化 move/rewrite apply。
5. 任一 revision 变化时停止相关操作。
6. 明确部分失败的恢复动作。
7. 执行后验证源文件、目标文件和相关链接。
8. 增加同名目标、外部修改和重复运行测试。

## 交付物

- 更新后的 `obsidian-inbox-triage/SKILL.md`
- plan/apply 示例
- 冲突与回滚场景测试

## 验收标准

- [ ] 推断出来的移动目标不会直接执行。
- [ ] 已存在目标不会被覆盖。
- [ ] 移动前后的链接一致。
- [ ] 冲突不会造成半整理状态。
- [ ] 重复执行不会重复移动或重复写 metadata。

## 验证

```bash
./scripts/lint-skills.sh
go test ./... -run 'Skill.*Inbox'
```

