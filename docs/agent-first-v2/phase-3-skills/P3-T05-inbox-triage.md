# P3-T05：迁移 Inbox Triage

- 状态：`已完成`
- 负责人：`Codex`
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

- [x] 推断出来的移动目标不会直接执行。
- [x] 已存在目标不会被覆盖。
- [x] 移动前后的链接一致。
- [x] 冲突不会造成半整理状态。
- [x] 重复执行不会重复移动或重复写 metadata。

## 验证

```bash
./scripts/lint-skills.sh
go test ./... -run 'Skill.*Inbox'
GOCACHE=/private/tmp/obs-cli-go-cache make release-check
```

## 完成记录

- Inbox Triage 已迁移为 V2 discover/read/plan/authorize/apply/verify 工作流，推断
  target、metadata 和批次子集必须授权。
- `note.move --dry-run` 新增确定性 `plan_hash`；apply 可用 `--if-plan-hash`
  绑定已审查的 source、target、链接改写集合和 revision。
- 成功 move 返回结构化 receipt，包含事务、请求、Vault、source/target revision
  及 `target_body_revision`。
- `note.get` 与 metadata get/set 提供 `body_revision`；Skill 记录逐字段 metadata
  revision chain，支持 partial 校验恢复和重复 no-op。
- `link.backlinks` 可检查不存在的旧 source；外部应用不遵守 CLI 锁时，通过
  apply→verify 静默窗口与写后复验发现 late backlink，并保守返回
  `partial/concurrent_external_change`。
- 场景测试覆盖既存目标、source 冲突、授权计划变化、链接事务、late external
  backlink、metadata partial/resume 与重复运行。
- 三轮独立前向测试完成；完整 release-check 通过，覆盖率 72.2%。
