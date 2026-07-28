# V1-T09：增加历史命名回流门禁

- 状态：`实现完成，待提交`
- 优先级：`高`
- 涉及项目：`obs-cli`
- 依赖：V1-T03～V1-T08

## 目标

建立自动门禁，阻止项目自有 V2 协议、配置、类型、函数、Capability、Schema 和当前文档
命名重新进入仓库，同时允许经过审计的第三方 module major version。

## 交付物

建议新增：

```text
scripts/naming-check.sh
```

并接入：

```text
Makefile: naming-check
release-check: naming-check
.github/workflows/ci.yml
```

## 检测范围

至少覆盖：

- `cmd/`
- `pkg/`
- `scripts/`
- `skills/`
- `testdata/`
- `docs/spec/`
- 当前 README 和命令文档
- `.github/`
- Makefile 和 GoReleaser 配置

## 禁止模式

至少检测：

```text
obs-cli/v2
v2.0.0-rc.1
config-v2.json
V2Config
NewV2Config
ValidateV2Config
renderV2
newV2Namespace
new*V2Command
daily_notes_v2
link_inspection_v2
metadata_v2
note_operations_v2
search_v2
*-v2.schema.json
```

## Allowlist 原则

允许：

- `go.mod`、`go.sum`、`vendor/` 中第三方 module major version。
- `THIRD_PARTY_NOTICES.md` 中第三方依赖的准确名称。
- 本任务组 `docs/v1-clean-start-refactor/` 中用于描述“被删除名称”的映射和验收文字。
  该例外只允许文档内容，不允许其 Schema、fixture 或代码被构建和运行时引用。

不允许：

- 通过整文件或整目录通配隐藏当前源码问题。
- 仅因为“测试需要”而允许旧项目协议；负向测试应使用集中 fixture 或构造字符串，并由
  精确行级/文件级 allowlist 说明目的。
- allowlist 包含 `cmd/`、`pkg/` 或当前规范目录。
- 以本任务组为跳板，让当前 README、规范或运行时继续引用旧名称。

## 执行步骤

1. 编写 shell 门禁，使用稳定、可读的模式列表。
2. 将第三方依赖路径与项目自有名称分开检测。
3. 对历史归档使用精确目录 allowlist。
4. 为负向协议测试提供最小 fixture allowlist。
5. 添加门禁自测：
   - 临时注入项目自有 `obs-cli/v2`，门禁失败；
   - 第三方 `tcell/v2` 不失败；
   - 删除注入后门禁通过。
6. 接入 Makefile。
7. 接入 PR CI 和 release-check。
8. 在门禁输出中打印文件、行号和命中模式。
9. 文档化新增 allowlist 的评审要求。

## 验收标准

- [ ] 当前仓库运行门禁通过。
- [ ] 人工注入每类禁止标记会失败。
- [ ] 第三方 `/v2` 不误报。
- [ ] allowlist 最小且有注释。
- [ ] 门禁进入本地 release-check 和 GitHub CI。
- [ ] 错误输出能直接定位文件和模式。
- [ ] 门禁不依赖网络。

## 验证命令

```bash
make naming-check
make format-check
make release-check
git diff --check
```

门禁自测必须使用临时目录或测试参数，不得把故意失败字符串提交到当前产品文件。

## 回滚

回滚门禁提交不会改变运行时，但会失去命名保护。T11 不允许在缺少该门禁时完成。

## 完成记录

- 完成日期：`2026-07-28（工作区实现）`
- 提交：`未提交`
- 禁止模式数量：`待填写`
- allowlist：`待填写`
- 失败注入证据：`待填写`
