# obs-cli v1.0.0-rc.1 联合候选验证记录

- 状态：验证通过，待人工批准真实 tag
- 验证日期：2026-07-28
- obs-cli 候选 commit：`631737543f0658bfff7944ac4d24aff58ea18acc`
- miniobsidian.nvim 配对 commit：`919c14fd16847d637d065e25d1056f4ec66566da`
- Plenary commit：`74b06c6c75e4eeb3108ec01852001636d85a932b`
- 说明：两个候选 commit 已冻结并通过联合验证；真实仓库未创建或推送
  `v1.0.0-rc.1` tag。

## 候选契约

```text
Product release:  v1.0.0-rc.1
CLI protocol:     obs-cli/v1
Config file:      config.json
Config schema:    1
Vault contract:   vault-contract/v1
Write protocol:   obs-write/v1
```

## 工具链

- Go：`go1.26.2 darwin/arm64`
- GoReleaser：`v2.17.1`（构建工具链 `go1.26.5`）
- Neovim：`v0.12.1`
- StyLua：`2.5.2`
- Selene：`0.28.0`

## 联合门禁

- `SKILL_EVAL_CLI_VERSION=v1.0.0-rc.1 make release-check`
  - Go vet、全部测试与 race 测试通过；
  - 覆盖率 `74.4%`，门槛 `70.0%`；
  - Schema、跨平台构建、许可证、Skill lint/eval 通过；
  - naming gate 及失败注入自测通过。
- `RC_VERSION=v1.0.0-rc.1 make rc-smoke`
  - 候选二进制 `--version` 为 `v1.0.0-rc.1`；
  - envelope 与 capabilities 只声明 `obs-cli/v1`；
  - Vault contract 为 `vault-contract/v1`；
  - Capability 不包含带历史版本后缀的 flag。
- `miniobsidian.nvim`
  - StyLua 检查通过；
  - Selene `0.28.0`：`0 errors / 0 warnings / 0 parse errors`；
  - Vault contract fixture gate 通过；
  - 完整 Plenary 测试通过；
  - 旧协议 mock 被判定为 `incompatible`。
- 开发态三入口 E2E：`6 / 6` 自动场景通过。
- GoReleaser 候选二进制 RC smoke：通过。
- GoReleaser 候选二进制三入口 E2E：`6 / 6` 自动场景通过。
- 全新配置目录只生成 `obs-cli/config.json`，且 `version == 1`。
- `TestDefaultStoreDoesNotReadOrMigrateHistoricalConfig` 与
  `TestStoreRejectsUnsupportedVersion` 通过。

miniobsidian 的首次沙箱内并行测试受 Neovim local socket 权限限制影响，文件监听断言失败；
在沙箱外使用相同 commit 和工具链完整复跑后全部通过。

## GoReleaser 制品验证

仓库内 `.goreleaser.yml` 已通过 `goreleaser check`。验证在 `/private/tmp` 的本地克隆中，
将临时 `v1.0.0-rc.1` tag 指向候选 commit，并使用
`goreleaser release --clean --skip=publish,announce` 构建。由于临时 tag 不存在于远端且
本地没有可用的私有仓库 GitHub 凭据，制品验证副本只把 changelog provider 从 `github`
改为 `git`；构建、版本注入、universal binary、归档、许可证和 checksum 配置未改变。

生成 6 个目标二进制、合并为 5 个归档：

| Artifact | SHA-256 |
|---|---|
| `obs-cli_darwin_all.tar.gz` | `e570c4b5addc47ae4feec4f584a4ef93977bd6c710e4055f4ebb43ca775315fc` |
| `obs-cli_linux_amd64.tar.gz` | `44dc247bab73f3819e62d3ce2511deb856818452c4e6fa588baec736bcff1603` |
| `obs-cli_linux_arm64.tar.gz` | `f5740c81357a1bc1f5af37c37e2e9992cc58db854fca10eab330a742165fa332` |
| `obs-cli_windows_amd64.tar.gz` | `4edd50a362d258224eea9a4af86e03fa804aa030eec96a2a6a9961bf338c9c0d` |
| `obs-cli_windows_arm64.tar.gz` | `afcffbc94dfcc2f3d93e7ce920a38890fe41554ab10c24d697f9ec972bec0663` |

`shasum -a 256 -c checksums.txt` 全部通过。macOS 归档包含 x86_64 + arm64 universal
binary；Linux 二进制为静态 ELF amd64/arm64；Windows 二进制为 PE32+ amd64/arm64。
归档包含 `LICENSE`、`THIRD_PARTY_NOTICES.md` 和 vendored license/NOTICE 文件。

## 验证期间修复的发布阻塞

- `9f5a002`：规范化 `go.mod` 的 direct/indirect 分类，确保 GoReleaser 的
  `before: go mod tidy` 不会制造 dirty tree。
- `2c90342`：ldflags 改用 `{{.Tag}}`，保留版本前缀 `v`；RC smoke 增加配置回归门禁。
- `6317375`：RC smoke 与三入口 E2E 支持直接验证打包后的候选二进制。

## 剩余人工动作

- 审阅本报告和配对 commit。
- 明确批准后，才可在 obs-cli 候选 commit 上创建并推送 annotated
  `v1.0.0-rc.1` tag。
- 当前验证不包含发布、tag 或 push，也未修改任何远端状态。
