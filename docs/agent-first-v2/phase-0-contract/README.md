# P0：共同协议与工程基线

## 阶段目标

在实现 V2 前冻结两个项目共同遵守的语义，避免 Go CLI 与 Lua 插件继续各自定义 Vault、Daily Note、Wikilink 和冲突行为。

## 进入条件

- 已确认采用破坏性 V2。
- 已确认 `miniobsidian.nvim` 不强制依赖 `obs-cli`。

## 任务进度

阶段进度：`6 / 6`

- [x] [P0-T01 产品边界与架构决策](./P0-T01-product-boundary.md)
- [x] [P0-T02 Vault 共同约定](./P0-T02-vault-conventions.md)
- [x] [P0-T03 CLI JSON 协议与错误模型](./P0-T03-cli-protocol.md)
- [x] [P0-T04 Revision、原子写入与冲突协议](./P0-T04-concurrency-contract.md)
- [x] [P0-T05 共享测试夹具](./P0-T05-shared-fixtures.md)
- [x] [P0-T06 开源许可与发布基线](./P0-T06-license-baseline.md)

推荐顺序：P0-T01 → P0-T02/P0-T03/P0-T06 → P0-T04 → P0-T05。

## 阶段完成标准

- [x] 所有共同规范均有版本号和变更规则。
- [x] 两个项目的 README 能链接到共同约定。
- [x] CLI JSON Schema 和稳定错误码完成评审。
- [x] 并发写入协议能覆盖 Obsidian、Agent、Neovim 同时编辑。
- [x] 共享 fixture 能被 Go 和 Lua 测试读取。
- [x] 发布物的许可证和派生关系处理方式明确。

## 阶段验证

```bash
cd /Users/andy/github/obs-cli
find docs/agent-first-v2 -type f -name '*.md' -print
find testdata/vault-contract -type f -print
```
