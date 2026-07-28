# V1-T11：验证 v1.0.0-rc.1 联合候选版本

- 状态：`验证中，待 GoReleaser artifact 与 tag 审批`
- 优先级：`发布阻断`
- 涉及项目：`obs-cli`、`miniobsidian.nvim`
- 依赖：V1-T10

## 目标

以两个已经通过联合门禁的精确 commit 生成并验证首个
`obs-cli v1.0.0-rc.1` 候选版本；在全部证据确认前不创建或推送 tag。

## 候选兼容集合

```text
obs-cli product:       v1.0.0-rc.1
obs-cli protocol:      obs-cli/v1
config file:           config.json
config schema:         1
write protocol:        obs-write/v1
vault contract:        vault-contract/v1
miniobsidian selector: 精确 commit 或首个明确的 V1 feature line
skill bundle:          agent-first-v1
```

## 发布前检查

```bash
git -C /Users/andy/github/obs-cli status --short
git -C /Users/andy/github/miniobsidian.nvim status --short
git -C /Users/andy/github/obs-cli rev-parse HEAD
git -C /Users/andy/github/miniobsidian.nvim rev-parse HEAD
git -C /Users/andy/github/obs-cli tag -l 'v1.0.0-rc.1'
```

要求：

- 两仓没有非预期未提交修改。
- T10 验证的 commit 与当前 HEAD 完全一致。
- tag 尚不存在。
- release notes、license、notices 和归档配置就绪。

## 执行步骤

1. 修复 Makefile 发布流程，使其不再尝试修改源码中的硬编码 Version。
2. 使用 ldflags/GoReleaser 注入 `v1.0.0-rc.1`。
3. 在临时目录生成六目标二进制和发布归档。
4. 对每个二进制验证：
   - `--version`；
   - `capabilities --output json`；
   - `protocol_versions == ["obs-cli/v1"]`；
   - `vault_contract == vault-contract/v1`；
   - 不包含旧 `_v2` flag。
5. 检查归档包含：
   - binary；
   - `LICENSE`；
   - `THIRD_PARTY_NOTICES.md`；
   - 要求的 vendored license/NOTICE；
   - checksums。
6. 使用候选二进制重新运行 RC smoke。
7. 使用候选二进制运行三入口 E2E，而不是开发态 `go run`。
8. 在全新配置目录验证 `config.json + version: 1`。
9. 验证只有 `config-v2.json` 时不存在隐式迁移。
10. 生成联合 RC 验证报告，记录 artifact checksum 和配对 commit。
11. 人工批准后才创建 annotated/pre-release tag。

## 发布命名

不得出现：

```text
v2.0.0-rc.1
agent-first-v2
obs-cli/v2
config-v2.json
```

候选发布说明应描述这是 Agent-first 产品的首次候选版本，不使用“从 V1 破坏性升级到
V2”的措辞。

## 验收标准

- [ ] 候选二进制版本为 `v1.0.0-rc.1`。
- [ ] 候选二进制只声明 `obs-cli/v1`。
- [ ] 新配置只创建 `config.json` 且 Schema 为 1。
- [ ] 发布归档和 checksums 完整。
- [ ] 候选二进制 RC smoke 通过。
- [ ] 候选二进制三入口 E2E 通过。
- [ ] 双仓配对 commit 已记录。
- [ ] release notes 无 V2 历史叙事。
- [ ] 在人工批准前未创建或推送 tag。

## 验证命令

```bash
SKILL_EVAL_CLI_VERSION=v1.0.0-rc.1 make release-check
RC_VERSION=v1.0.0-rc.1 make rc-smoke
make naming-check
./scripts/run-three-client-e2e.sh
git diff --check
```

发布归档验证命令由 GoReleaser 配置和 release workflow 固化，并将输出写入验证报告。

## 回滚

tag 创建前：删除临时 artifact，按任务提交回滚并完整重跑 T10。

tag 创建后：不得移动或重写 tag。若发现问题，撤下/标记该 pre-release，并通过新的
`v1.0.0-rc.N` 修复。

## 完成记录

- 完成日期：`待填写`
- obs-cli commit：`待填写`
- miniobsidian.nvim commit：`待填写`
- Artifact：`待填写`
- Checksums：`待填写`
- RC smoke：`待填写`
- Three-client E2E：`待填写`
- Tag 审批：`待填写`
