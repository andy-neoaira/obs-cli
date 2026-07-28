# obs-cli V1 首版规范化重构计划

- 状态：`11 项任务完成，RC 验证通过，待人工批准 tag`
- 计划类型：一次性协调硬切换
- 首个候选版本：`v1.0.0-rc.1`
- CLI 协议：`obs-cli/v1`
- 配置 Schema：`1`
- Vault 契约：`vault-contract/v1`
- 写入协议：`obs-write/v1`
- 涉及项目：
  - `/Users/andy/github/obs-cli`
  - `/Users/andy/github/miniobsidian.nvim`

## 1. 决策

Agent-first 方向是当前产品的第一版，不继承旧 Human-first CLI 的版本叙事。本次重构
把现有尚未发布的 V2 工作线重新定义为正式 V1，并一次性清除源码、协议、配置、
Schema、Capability、Skill、测试和当前产品文档中的历史版本包袱。

本计划采用硬切换，不提供过渡兼容：

- 不同时支持 `obs-cli/v1` 与 `obs-cli/v2`。
- 不保留 `V2Config`、`newNoteV2Command`、`renderV2` 等兼容别名。
- 不同时读取 `config.json` 与 `config-v2.json`。
- 不为旧 feature flag 提供别名。
- 不在正式运行时代码中加入旧配置自动迁移。
- 不把旧阶段文档继续作为当前产品事实源。
- 不修改第三方 Go module 自身的 `/v2` 路径。

Git 历史、许可证和 `THIRD_PARTY_NOTICES.md` 已足够保存派生关系与开发历史。当前实现
不需要通过命名继续承载这些历史。

## 2. 目标状态

完成后，当前产品表面必须统一为：

```text
Product release:  v1.0.0-rc.1
CLI protocol:     obs-cli/v1
Config file:      config.json
Config schema:    1
Vault contract:   vault-contract/v1
Write protocol:   obs-write/v1
```

Go 内部代码只使用无版本业务名称：

```text
Config
ConfigPath
Registry
newNoteCommand
newVaultCommand
newNamespace
renderEnvelope
```

所有需要版本化的外部契约从 `v1` 开始。内部实现文件、类型和函数不带版本号，除非未来
确实需要两个版本同时存在，并通过新的架构决策批准。

## 3. 成功标准

- 当前产品源码、测试、Skill、脚本和用户文档中不存在项目自有的 V2 命名。
- `obs-cli` 只产生 `obs-cli/v1` envelope。
- `capabilities` 只声明 `obs-cli/v1`。
- `miniobsidian.nvim` 只接受 `obs-cli/v1`。
- 新配置只使用 `config.json`，内容 `version` 固定为 `1`。
- 所有首版 JSON Schema 文件和 `$id` 使用 `-v1`。
- Capability 不再包含 `_v2` 后缀。
- 当前文档不再把新产品描述为从 V1 升级到 V2。
- 自动门禁能阻止项目自有 V2 命名重新进入仓库。
- 两个仓库的 CI、契约测试和三入口 E2E 全部通过。
- `v1.0.0-rc.1` 联合候选版本验证通过后才允许创建 tag。

## 4. 非目标

- 不为尚未公开发布的 `obs-cli/v2` 提供兼容期。
- 不保证当前开发机上的 `config-v2.json` 自动可用。
- 不保留旧 CLI 命令、旧 TTY picker 或 GUI 启动行为。
- 不修改 `vault-contract/v1`、`obs-write/v1`、Agent Result V1 或 Agent Handoff V1
  的既有语义。
- 不重写第三方依赖的 module major version。
- 不在本阶段增加新的业务 operation。
- 不借命名重构改变 Note、Vault、Daily、Search、Link 或 Metadata 的业务语义。

## 5. 执行原则

1. 先冻结命名，再修改实现，禁止边改边决定名称。
2. 协议、Schema、fixture、消费者必须在同一任务链中同步。
3. 一个任务对应一个可独立审计的提交；跨仓库任务分别提交，但记录配对 commit。
4. 不引入 deprecated alias、fallback、dual-read 或 compatibility shim。
5. 每个任务先记录基线和用户已有改动；本次执行已确认
   `miniobsidian.nvim` 起始工作区干净。
6. 机械改名和行为修改分开；本计划原则上只做命名、版本基线和契约一致性重构。
7. 发现真实行为缺陷时另建任务，不混入批量改名提交。
8. 所有任务完成后再生成 tag；中间提交不对外宣称兼容或可发布。

## 6. 任务依赖

```text
V1-T01 命名冻结
   │
   ├──> V1-T02 产品与协议基线
   │       │
   │       ├──> V1-T03 Go 内部命名
   │       ├──> V1-T04 配置重置
   │       └──> V1-T05 Schema/fixture/报告
   │
   ├──> V1-T06 Capability 清理
   │
   └──────────────────────────────┐
                                  v
                         V1-T07 miniobsidian 同步
                                  │
                         V1-T08 历史文档收口
                                  │
                         V1-T09 命名回流门禁
                                  │
                         V1-T10 联合回归与 E2E
                                  │
                         V1-T11 首次 RC 验证
```

T03、T04、T05、T06 可在 T01、T02 完成后并行，但在 T07 开始前必须合并并形成稳定的
CLI 契约。

## 7. 任务清单

实施进度：`11 / 11`。真实 tag 和 push 是验证后的独立人工审批动作，不属于本次自动执行。

- [x] [V1-T01 冻结命名规范和映射表](./V1-T01-naming-contract.md)
- [x] [V1-T02 重置产品版本和 CLI 协议](./V1-T02-product-protocol-v1.md)
- [x] [V1-T03 清理 Go 内部 V2 命名](./V1-T03-go-symbol-cleanup.md)
- [x] [V1-T04 重置配置文件与 Schema 版本](./V1-T04-config-v1-reset.md)
- [x] [V1-T05 重命名 Schema、Golden Fixture 和 Skill 报告](./V1-T05-schema-fixture-skill-reports.md)
- [x] [V1-T06 清理 Capability 版本后缀](./V1-T06-capability-cleanup.md)
- [x] [V1-T07 同步 miniobsidian.nvim Adapter](./V1-T07-miniobsidian-sync.md)
- [x] [V1-T08 删除历史 V2 文档并收口当前文档](./V1-T08-history-doc-cleanup.md)
- [x] [V1-T09 增加历史命名回流门禁](./V1-T09-naming-regression-gate.md)
- [x] [V1-T10 运行双仓 CI 与三入口 E2E](./V1-T10-joint-validation.md)
- [x] [V1-T11 验证 v1.0.0-rc.1 联合候选版本](./V1-T11-first-rc-validation.md)

当前验证摘要（2026-07-28）：

- `obs-cli`: `make release-check` 通过，覆盖率 `74.4%`，RC smoke 为
  `v1.0.0-rc.1`。
- `miniobsidian.nvim`: StyLua、Selene、fixture gate、完整 Plenary 测试通过。
- 开发态与 GoReleaser 候选二进制三入口 E2E：均为 `6 / 6` 通过。
- GoReleaser `v2.17.1` 配置校验、5 个归档、checksum 和候选二进制 RC smoke 通过。
- 候选配对：obs-cli `6317375`，miniobsidian.nvim `919c14f`。
- 未创建或推送真实 `v1.0.0-rc.1` tag。

## 8. 跨仓库提交策略

推荐提交顺序：

1. `obs-cli`：T01 文档决策。
2. `obs-cli`：T02～T06 实现与契约重构。
3. `miniobsidian.nvim`：T07 Adapter、测试和文档同步。
4. `obs-cli`：T08～T09 文档与门禁收口。
5. 两仓分别提交 T10 验证所需修正。
6. `obs-cli`：T11 发布验证记录。

T07 开始前必须记录两个仓库的：

```bash
git status --short
git branch --show-current
git rev-parse HEAD
```

执行前确认 `miniobsidian.nvim` 工作区干净；任务只修改并提交了计划列出的 9 个文件。

## 9. 通用验证

每个 `obs-cli` 任务至少运行：

```bash
gofmt -w <本任务修改的 Go 文件>
go test ./相关包/...
git diff --check
```

阶段收口运行：

```bash
make release-check
./scripts/run-three-client-e2e.sh
```

`miniobsidian.nvim` 使用其仓库固定的格式、静态检查、Plenary 测试和共享 fixture 门禁。
具体命令以 T07、T10 为准。

## 10. 回滚原则

本计划不支持运行时兼容回滚。代码回滚只能以任务提交为单位进行：

- T02～T06 任一任务回滚后，不能运行新 Adapter。
- T07 回滚后，必须同时回滚到旧 CLI commit，不能混用 `/v1` 与 `/v2`。
- T04 回滚会改变配置路径，必须使用测试配置目录验证，不能让程序猜测旧文件。
- T11 创建 tag 前允许按提交回滚；tag 创建后必须通过新的正式版本修复，不改写 tag。

## 11. 完成定义

只有满足以下全部条件，本计划才能标记完成：

- 11 个任务全部完成并记录提交。
- 两仓工作区没有非预期修改。
- 禁止命名门禁无误报且能捕获人工注入的 V2 标记。
- 两仓 CI、race、Schema、Skill eval、共享 fixture 和三入口 E2E 通过。
- `capabilities`、实际 envelope、Schema、Adapter 和文档均只声明 `obs-cli/v1`。
- 候选版本输出为 `v1.0.0-rc.1`。
- 未创建任何兼容别名、双读逻辑或隐式迁移路径。
