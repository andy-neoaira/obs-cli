# V1-T05：重命名 Schema、Golden Fixture 和 Skill 报告

- 状态：`实现完成，待提交`
- 优先级：`高`
- 涉及项目：`obs-cli`
- 依赖：V1-T01、V1-T02

## 目标

将当前首版公共 Schema、测试 golden fixture 和 Skill 报告统一为 V1 文件名、`$id`
和引用，删除所有 `*-v2` 当前产品资产。

## 强制映射

```text
response-v2.schema.json                 -> response-v1.schema.json
capabilities-v2.schema.json             -> capabilities-v1.schema.json
search-audit-report-v2.schema.json      -> search-audit-report-v1.schema.json
compare-synthesis-report-v2.schema.json -> compare-synthesis-report-v1.schema.json
project-status-report-v2.schema.json    -> project-status-report-v1.schema.json
cmd/testdata/*-v2.json                  -> cmd/testdata/*-v1.json
```

## 修改范围

- `docs/spec/schema/`
- `cmd/testdata/`
- `testdata/protocol/`
- `cmd/*schema*test.go`
- Skill `SKILL.md`
- Skill eval manifest、schema 和报告
- `scripts/schema-check.sh`
- 当前规范文档

## 禁止事项

- 不保留旧 Schema 文件作为 redirect。
- 不在测试中同时加载 V1 和 V2。
- 不保留旧 `$id`。
- 不创建内容相同但名称不同的重复 Schema。
- 不借机放宽 `additionalProperties`、required 或错误码约束。

## 执行步骤

1. 建立旧文件到新文件的完整清单。
2. 使用文件移动保留 Git rename 可读性。
3. 更新每个 Schema 的 `$id`、title 和协议 const。
4. 更新所有 Go 测试中的相对路径。
5. 更新 golden fixture 文件名和内部协议值。
6. 更新 Skill 文档中的报告 Schema 引用。
7. 更新 `scripts/schema-check.sh`。
8. 更新 eval manifest 和 release report。
9. 运行 Schema 合同测试和 Skill 场景测试。
10. 检索孤立旧文件名和旧 `$id`。

## 验收标准

- [ ] 当前 Schema 均使用 `-v1.schema.json`。
- [ ] `$id` 与实际文件名一致。
- [ ] golden fixture 使用 `-v1.json`。
- [ ] 所有 Skill 引用 V1 报告 Schema。
- [ ] Schema 内容约束未被意外放宽。
- [ ] schema-check、Skill lint 和 Skill eval 通过。
- [ ] 不存在重复 V2 Schema。

## 验证命令

```bash
make schema-check
make skill-lint
make skill-evals
go test ./cmd ./pkg/protocol
find docs/spec/schema cmd/testdata -type f -iname '*v2*' -print
rg -n -- '-v2\.schema\.json|-v2\.json|obs-cli/v2' \
  cmd pkg skills scripts testdata docs/spec
git diff --check
```

## 回滚

Schema 文件移动、引用修改和 fixture 移动必须作为同一提交回滚。不得让默认分支出现
Schema 已移动但测试仍引用旧路径的中间状态。

## 完成记录

- 完成日期：`2026-07-28（工作区实现）`
- 提交：`未提交`
- Schema 数量：`待填写`
- Golden fixture 数量：`待填写`
- Skill 引用数量：`待填写`
