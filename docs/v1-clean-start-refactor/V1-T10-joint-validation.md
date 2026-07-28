# V1-T10：运行双仓 CI 与三入口 E2E

- 状态：`双仓自动回归通过，待配对提交冻结`
- 优先级：`阻断`
- 涉及项目：`obs-cli`、`miniobsidian.nvim`
- 依赖：V1-T02～V1-T09

## 目标

验证硬切换后的 CLI、Skills、Neovim Adapter 和共享 Vault 契约形成一个内部一致的 V1
兼容集合，并证明不存在旧协议 fallback。

## 执行前冻结

记录：

```text
obs-cli commit
miniobsidian.nvim commit
Go version
Neovim version
Plenary commit/path
操作系统和架构
```

两个仓库必须使用将要进入 RC 的精确 commit，不允许测试后继续修改而不重跑。

## obs-cli 门禁

必须运行：

```bash
make format-check
make naming-check
make compatibility-check
go vet ./...
go test ./...
go test -race ./...
make coverage-check
make schema-check
make build-check
make license-check
make skill-lint
make skill-evals
make rc-smoke
```

推荐直接运行：

```bash
make release-check
```

并保存每个子门禁结果，不能只记录最终 exit code。

## miniobsidian.nvim 门禁

按仓库 CI 固定版本运行：

- StyLua；
- Selene；
- 全部 Plenary 测试；
- Vault contract fixture；
- CLI Adapter 测试；
- Move、Handoff、External Changes 测试；
- 文档 tag 生成一致性。

不得用系统上任意版本的 Neovim 替代仓库声明版本而不记录差异。

## 三入口 E2E

运行：

```bash
./scripts/run-three-client-e2e.sh
```

至少验证：

1. Obsidian 风格外部文件事件后 CLI 能读取最新 revision。
2. Agent CLI create/patch/move 使用 `obs-cli/v1`。
3. Neovim Adapter 接受 `/v1` 并拒绝 `/v2`。
4. stale revision 返回 `REVISION_CONFLICT`。
5. move dry-run 和 plan hash 一致。
6. Daily Note 路径与模板契约一致。
7. dirty buffer 不被 CLI 结果静默覆盖。
8. 共享 `vault-contract/v1` fixture hash 一致。

## 负向联合测试

- 用旧协议 mock 启动 Adapter，必须 incompatible。
- 在配置目录只放 `config-v2.json`，CLI 不得自动加载。
- capability 中注入旧 `_v2` flag，命名门禁必须失败。
- 使用 `version: 2` 配置，必须明确拒绝。
- 使用 V2 Schema 文件名，schema/naming gate 必须失败。

## 验收标准

- [ ] `obs-cli make release-check` 通过。
- [ ] `miniobsidian.nvim` 完整 CI 通过。
- [ ] 三入口 E2E 全部通过。
- [ ] 所有负向硬切换测试通过。
- [ ] 两仓共享 fixture 一致。
- [ ] 工作区没有测试生成的非预期文件。
- [ ] 验证后 HEAD 未变化。
- [ ] 结果记录包含精确 commit 和工具版本。

## 失败处理

- 命名/引用失败：返回对应 T02～T09 修复。
- 业务行为失败：新建独立缺陷任务，不在 T10 顺手修改。
- 工具链缺失：修复执行环境后重跑，不把环境失败记为通过。
- 任一代码修改发生后：完整重跑受影响仓库和三入口 E2E。

## 回滚

T10 不修改业务代码，只记录验证证据。若发现问题，回滚对应任务提交并重新建立配对
commit，不允许发布部分通过的组合。

## 完成记录

- 完成日期：`待配对提交冻结后填写`
- obs-cli commit：`基线 01be5bf3eb36b5417465833ec5d897fe2e403cc7 + 未提交工作区`
- miniobsidian.nvim commit：`基线 184662ddddadf6fe887d8897e5874c37779757c2 + 未提交工作区`
- 工具版本：`Go 1.26.2；Neovim 0.12.1；Selene 0.28.0`
- obs-cli release-check：`通过；coverage 74.4%`
- miniobsidian CI：`StyLua、Selene、fixture gate、完整 Plenary 通过`
- three-client E2E：`6 / 6 通过`
- 负向测试：`旧协议、历史配置版本、历史 Capability 名称门禁均通过`
