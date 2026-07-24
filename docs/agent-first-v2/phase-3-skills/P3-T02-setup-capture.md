# P3-T02：迁移 Vault Setup 与 Capture

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`obs-cli Skills`
- 依赖：P3-T01、P1-T01

## 目标

迁移 Vault 初始化和快速捕获两个基础 Skill 到 V2 协议。

## 实施步骤

1. 更新 `obsidian-vault-setup` 使用 `vault discover/list/add/set-default`。
2. 明确修改默认 Vault 属于配置写入，执行前展示变化。
3. 更新 `obsidian-capture` 使用 `note create/append`。
4. 捕获前明确目标 Vault、路径和模式。
5. 长内容通过 stdin 传输。
6. create 冲突时不自动 overwrite。
7. apply 后重新读取目标并验证 revision/摘要。
8. 增加重复请求的幂等场景。

## 交付物

- 更新后的两个 `SKILL.md`
- setup/capture 场景测试

## 验收标准

- [ ] 不调用 V1 命令。
- [ ] 不写入 Obsidian 官方配置。
- [ ] 捕获不会静默覆盖已有笔记。
- [ ] 结果包含 Vault、路径、操作类型和新 revision。
- [ ] capability 不满足时给出明确升级提示。

## 验证

```bash
./scripts/lint-skills.sh
go test ./... -run 'Skill.*(Setup|Capture)'
```

