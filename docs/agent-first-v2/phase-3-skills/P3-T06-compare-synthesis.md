# P3-T06：新增 Compare 与 Knowledge Synthesis

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli Skills`
- 依赖：P3-T01、P3-T04

## 目标

提供多笔记比较与知识综合两个核心 Agent 读取分析场景。

## 实施步骤

1. 创建 `obsidian-compare-notes`。
2. 支持显式笔记列表、搜索结果集和同主题笔记比较。
3. 比较结构、Frontmatter、事实、观点、TODO 和更新时间。
4. 创建 `obsidian-knowledge-synthesis`。
5. 综合结果保留来源路径和 revision。
6. 默认只输出报告；写回新笔记必须进入 dry-run/apply。
7. 定义来源变化时综合结果过期的提示。
8. 增加冲突信息、重复来源和无足够证据场景。

## 交付物

- 两个新 Skill
- 比较/综合结果格式
- 场景与黄金输出

## 验收标准

- [x] 分析结论可追溯到来源笔记。
- [x] 不把推断伪装成原笔记事实。
- [x] 默认不修改来源笔记。
- [x] 写回时使用 create/patch 和 revision。
- [x] 来源内容变化能被识别。

## 验证

```bash
./scripts/lint-skills.sh
go test ./... -run 'Skill.*(Compare|Synthesis)'
GOCACHE=/private/tmp/obs-cli-go-cache make release-check
```

## 完成记录

- 新增 `obsidian-compare-notes`：支持显式/搜索来源、规范 path 去重、结构与
  metadata/事实/观点/TODO/modified_at 比较，默认严格只读。
- 新增 `obsidian-knowledge-synthesis`：保留来源 statement、综合、推断、冲突和
  gaps；只有明确 writeback 授权才使用 create 或唯一 anchor patch。
- `note.get` 新增稳定快照 `modified_at`；与 revision/body_revision 一起形成来源
  manifest，verify 变化时重算一次或安全返回 stale。
- 新增 compare/synthesis 共享报告 Schema，覆盖 selection、source/evidence/claim
  引用、staleness、writeback digest/request ID 和 unknown outcome。
- CI 引入 vendored Draft 2020-12 validator，黄金输出和危险负例都执行真实 Schema
  校验；另有跨来源/唯一 ID/path/revision/modified_at 的语义校验。
- 场景测试覆盖来源变化、mtime-only 变化、create 竞态与 exact no-op、patch revision
  冲突、anchor 歧义、dry-run 后来源变化停止写入。
- 两轮独立前向测试完成；排序质量、连续 revision 抖动、超时注入留给 P3-T08。
- 完整 release-check 通过，覆盖率 72.3%。
