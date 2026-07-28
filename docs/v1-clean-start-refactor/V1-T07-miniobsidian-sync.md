# V1-T07：同步 miniobsidian.nvim Adapter、测试和文档

- 状态：`实现与完整验证完成，待提交`
- 优先级：`阻断`
- 涉及项目：`miniobsidian.nvim`、`obs-cli`
- 依赖：V1-T02、V1-T05、V1-T06

## 目标

让 `miniobsidian.nvim` 只接受 `obs-cli/v1`，同步产品版本、测试 fixture 和用户文档，
并保持插件对 CLI 的可选依赖边界。

## 已知基线

主要运行时耦合位于：

```text
/Users/andy/github/miniobsidian.nvim/lua/miniobsidian/cli.lua
```

当前直接引用还分布于：

- `tests/cli_spec.lua`
- `tests/move_spec.lua`
- `tests/handoff_spec.lua`
- `tests/three_client_e2e_spec.lua`
- `README.md`
- `README.en.md`
- `doc/miniobsidian.txt`
- `doc/miniobsidian.zh.txt`
- `MANUAL_TESTING.md`

执行前 `miniobsidian.nvim` 存在用户未提交修改，必须保留。

## 禁止事项

- 不同时接受 `obs-cli/v1` 与 `obs-cli/v2`。
- 不增加协议 fallback。
- 不降低 protocol mismatch 的错误级别。
- 不让插件变成强制依赖 CLI。
- 不覆盖当前未提交修改。
- 不使用整文件复制覆盖冲突文件。
- 不修改 `vault-contract/v1`。

## 执行前检查

```bash
git -C /Users/andy/github/obs-cli status --short
git -C /Users/andy/github/obs-cli rev-parse HEAD
git -C /Users/andy/github/miniobsidian.nvim status --short
git -C /Users/andy/github/miniobsidian.nvim rev-parse HEAD
git -C /Users/andy/github/miniobsidian.nvim diff -- \
  lua/miniobsidian/cli.lua tests/cli_spec.lua tests/move_spec.lua \
  tests/handoff_spec.lua tests/three_client_e2e_spec.lua
```

若目标行和用户改动重叠，先记录冲突并使用最小局部 patch。

## 执行步骤

1. 将 Adapter 的唯一协议常量改为 `obs-cli/v1`。
2. 更新 protocol mismatch 错误消息和 expected/actual details。
3. 确认 capabilities 仍按 operation 建立可用能力集合。
4. 若 T06 删除 feature flag，确认插件没有隐式依赖。
5. 更新 CLI、Move、Handoff、三入口 E2E fixture：
   - protocol；
   - CLI candidate version；
   - capabilities 协议列表。
6. 增加负向测试：
   - `/v2` envelope 被拒绝；
   - capabilities 只声明 `/v2` 时状态为 incompatible；
   - `/v1` + 正确 Vault contract 才进入 ready。
7. 更新中英文 README、Vim help 和手工测试文档。
8. 重新生成 `doc/tags`，不得手工留下过时 tag。
9. 运行格式、静态检查、Plenary 测试。
10. 记录与 `obs-cli` 配对 commit。

## 测试矩阵

| CLI 状态 | Adapter 预期 |
|---|---|
| 未安装 | `unavailable`，插件本地功能可用 |
| 禁用 | `disabled` |
| 合法 `obs-cli/v1` | `ready` |
| `obs-cli/v2` | `incompatible` |
| Vault contract 不匹配 | `incompatible` |
| 缺少 operation | 对应工作流不可用 |
| 非法 JSON | `error` |

## 验收标准

- [ ] Adapter 唯一协议常量为 `obs-cli/v1`。
- [ ] `/v2` 不被接受。
- [ ] 可选 CLI 边界不变。
- [ ] operation capability gate 不回归。
- [ ] Move、Handoff 和三入口 fixture 同步。
- [ ] 中英文 README、help 和手工测试一致。
- [ ] 用户原有未提交修改完整保留。
- [ ] 插件完整测试通过。

## 验证命令

在 `miniobsidian.nvim` 仓库按现有 CI 使用固定工具链运行：

```bash
stylua lua plugin tests
selene lua
make test
git diff --check
rg -n 'obs-cli/v2|v2\.0\.0-rc\.1' \
  lua tests README.md README.en.md doc MANUAL_TESTING.md
```

若仓库的 `make test` 需要固定 Neovim/Plenary 路径，使用 CI 文档声明的等价命令。检索必须
无输出。

## 回滚

回滚 `miniobsidian.nvim` 配对提交，并同时将运行环境切回对应的旧 CLI commit。不得只
回滚 Adapter 协议常量而保留新 fixture 或文档。

## 完成记录

- 完成日期：`2026-07-28（工作区实现）`
- obs-cli 配对提交：`未提交`
- miniobsidian.nvim 提交：`未提交`
- 原有 dirty 文件保护说明：`执行前工作区干净；仅修改任务列出的 9 个目标文件`
- 测试摘要：`StyLua、Selene、fixture gate、完整 Plenary 测试通过`
