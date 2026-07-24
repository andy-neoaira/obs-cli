# P0-T05：共享测试夹具

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli`、`miniobsidian.nvim`
- 依赖：P0-T02、P0-T04

## 目标

建立两个实现共同使用的 Vault 样例和黄金结果，防止相同笔记在 Go 与 Lua 中得到不同解释。

## 实施步骤

1. 在 `obs-cli/testdata/vault-contract/` 建立版本化 fixture。
2. 覆盖基础 Vault、Unicode、空格、同名笔记、坏链接、Daily Template。
3. 覆盖符号链接逃逸、`..`、绝对路径和大小写冲突。
4. 覆盖 Frontmatter、CRLF、无结尾换行和空文件。
5. 覆盖 concurrent update 的原始内容、外部修改和预期冲突。
6. 为每个 fixture 编写 `expected.json`。
7. 为 `miniobsidian.nvim` 提供同步脚本或固定版本副本，禁止手工漂移。
8. 在 CI 中校验 fixture 版本和内容摘要。

## 交付物

- `testdata/vault-contract/v1/`
- `testdata/vault-contract/v1/manifest.json`
- fixture 同步或校验脚本

## 验收标准

- [x] 每个边界案例都有输入和机器可读预期。
- [x] Fixture 不包含真实个人笔记或隐私数据。
- [x] Go 与 Lua 验证使用相同版本。
- [x] 同步脚本的 `--check` 能检测两个本地仓库的未同步状态。
- [x] 所有实际检入路径均可跨平台检出；Windows 非法名称使用 JSON 逻辑输入表达。

## 验证

```bash
go test ./... -run ContractFixture
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/contract" -c qa
```

## 验证记录

- 2026-07-24：建立 9 组 fixture、32 个文件，全部 JSON 通过语法检查。
- 2026-07-24：Go manifest 测试验证版本、case、expected 与隐私标记。
- 2026-07-24：Neovim Lua 验证同步副本的 9 个 case 均为 `vault-contract/v1`。
- 2026-07-24：`sync-vault-contract-fixtures.sh --check` 通过，两个本地副本一致。
- 说明：跨仓库 CI 摘要门禁在 P2-T01 建立插件 CI 时接入；P0 交付可执行的同步与检查机制，避免依赖尚不存在的 P2 CI。
