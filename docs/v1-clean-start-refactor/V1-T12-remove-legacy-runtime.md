# V1-T12：删除旧 Human-first 运行时

- 状态：`完成`
- 优先级：`高`
- 涉及项目：`obs-cli`、`miniobsidian.nvim`
- 依赖：V1-T01～V1-T11

## Review 结论

`pkg/actions` 不是当前 V1 命令实现，而是重构前 Human-first CLI 的完整业务入口。当前
`cmd` 没有生产调用者引用它；该包只被自身测试继续保活。

旧运行时包含：

- 依赖 TTY 的 fuzzy picker；
- 自动启动编辑器或 Obsidian URI；
- `NoteManager`、`VaultManager`、`UriManager`、`FuzzyFinderManager` 等旧 façade；
- 旧式 create、daily、delete、frontmatter、list、move、open、print、search action；
- 只为上述代码服务的 mocks、测试和第三方交互依赖。

这些能力与 ADR-001 冻结的 Agent-first 边界冲突，不属于当前 V1，也不承担仍在使用的
协议兼容职责。

`miniobsidian.nvim` 只消费 `obs-cli/v1`，没有双协议运行路径。插件中发现的
`follow_link()` 与 `follow_link_or_gf()` 是仅为旧用户配置避免报错而保留的空函数，
当前代码没有调用者，属于明确的 compatibility shim。

## 决策

采用硬删除，不提供 deprecated alias、迁移层或旧命令转发：

1. 删除整个 `pkg/actions`。
2. 删除只为该包服务的 mocks 和旧 `pkg/obsidian` façade。
3. `pkg/obsidian` 仅保留当前 V1 仍使用的纯解析、发现和结构化链接重写能力。
4. 删除 fuzzy picker、外部应用启动器及其传递依赖，重新生成 vendor。
5. 删除 `miniobsidian.nvim` 的两个旧配置空 API。
6. 增加架构门禁，禁止旧目录、旧 Manager 接口和旧交互依赖回流。

## 当前边界

```text
cmd
 └─ noteops.Service / vault Registry
     ├─ pathpolicy
     ├─ storage（revision、原子写入、事务）
     └─ obsidian（当前格式解析与结构化链接处理）
```

`obs-cli` 不再包含另一套并行的 action/manager 业务栈。`miniobsidian.nvim` 仍是独立
Neovim 客户端；其可选 CLI adapter 是当前产品边界，不是旧版本兼容层。

## 验收标准

- [x] `pkg/actions` 与旧 façade 已删除。
- [x] 当前生产代码不存在旧 Manager 接口或旧 action import。
- [x] fuzzy picker、外部应用启动器及其传递依赖已从 module/vendor/notices 删除。
- [x] `miniobsidian.nvim` 的旧配置空 API 已删除。
- [x] 自动门禁覆盖旧目录、旧符号和旧依赖。
- [x] `obs-cli make release-check` 通过，覆盖率 `72.7%`。
- [x] `miniobsidian.nvim make ci` 通过。
- [x] 三入口 E2E `6 / 6` 通过。

## 验证命令

```bash
make architecture-check
make release-check
./scripts/run-three-client-e2e.sh
```

`miniobsidian.nvim` 使用其固定的 StyLua、Selene、fixture 和 Plenary CI。

## 完成记录

- 完成日期：`2026-07-29`
- `miniobsidian.nvim` 配对提交：`6675bdb`
- `obs-cli`：format、命名、架构、兼容矩阵、vet、test、race、coverage、Schema、
  跨平台构建、许可证、Skill 和 RC smoke 全部通过。
- `miniobsidian.nvim`：StyLua、Selene、fixture 与完整 Plenary 测试通过。
- 三入口 E2E：`6 / 6` 通过。

## 回滚原则

本任务不支持运行时兼容回滚。如果未来确实需要人类交互层，应在当前 `cmd` 和
`noteops.Service` 之上设计新的外围 adapter，并通过新的 ADR 批准；不得恢复旧 action
或 manager 栈。
