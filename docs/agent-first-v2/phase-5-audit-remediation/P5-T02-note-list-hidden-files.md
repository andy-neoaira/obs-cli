# P5-T02：Note List 隐藏文件隔离

- 状态：`已完成`
- 优先级：`高`
- 负责人：`待分配`
- 涉及项目：`obs-cli`
- 依赖：P1-T02、P1-T06
- 来源：审计 M2（复核后修正问题描述）

## 背景与已确认现象

`noteops.Service.List` 会跳过隐藏目录，但不会提前跳过 `.draft.md` 等隐藏
Markdown 文件。文件随后被 `pathpolicy.Resolver` 拒绝，导致整个 `note list`
返回 `PATH_OUTSIDE_VAULT`；它并不会成功泄露隐藏文件。

## 目标

默认列表忽略任何隐藏路径段，同时保留 Resolver 作为防御性边界校验。

## 非目标

- 不增加读取隐藏文件的 capability。
- 不改变普通 Markdown 排序、扩展名大小写或符号链接策略。
- 不把隐藏文件作为错误返回给 Agent。

## 修改范围

- `pkg/noteops/service.go`
- `pkg/noteops/service_test.go` 或新的 list 专项测试
- `docs/spec/NOTE_OPERATIONS.md`

## 执行过程

1. 增加失败测试：Vault 根目录包含 `.draft.md` 时，当前 List 返回错误。
2. 增加嵌套隐藏文件、隐藏目录和普通文件组合 fixture。
3. WalkDir 遇到任意 `.` 前缀条目时：
   - 目录返回 `filepath.SkipDir`；
   - 文件、链接及其他条目返回 `nil`。
4. 普通 Markdown 仍通过 Resolver 后才加入结果。
5. 增加显式隐藏 scope 被 `pathpolicy` 拒绝的回归测试。
6. 更新 Note List 规范，明确默认不暴露隐藏路径。

## 验收标准

- [x] `.draft.md` 不出现在结果中，也不会导致 List 失败。
- [x] `Folder/.private.md` 与 `.obsidian/**` 均被忽略。
- [x] 隐藏文件存在时，所有普通笔记仍按稳定顺序返回。
- [x] 显式访问隐藏路径仍返回 `PATH_OUTSIDE_VAULT`。
- [x] 不跟随目录符号链接，现有路径安全测试保持通过。

## 验证命令

```bash
go test ./pkg/noteops -run 'List|Hidden'
go test ./pkg/pathpolicy ./pkg/noteops
go test -race ./pkg/noteops
make release-check
git diff --check
```

## 回滚

回滚独立提交即可；回滚后保留失败测试证据到任务记录，避免未来再次把
“隐藏文件导致全局失败”误判为预期行为。

## 完成记录

- 完成日期：`2026-07-28`
- 实现：WalkDir 对所有 `.` 前缀条目执行统一隔离；隐藏目录跳过子树，隐藏文件
  和链接直接忽略，普通 Note 继续通过 Resolver。
- 测试：覆盖根目录、嵌套隐藏文件、`.obsidian`、隐藏子树和显式隐藏读取拒绝。
- 验证：定向测试、`pkg/noteops` race 与 `make release-check` 通过；总覆盖率
  `72.8%`。
