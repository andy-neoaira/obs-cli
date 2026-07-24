# P3-T06：新增 Compare 与 Knowledge Synthesis

- 状态：`未开始`
- 负责人：`待分配`
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

- [ ] 分析结论可追溯到来源笔记。
- [ ] 不把推断伪装成原笔记事实。
- [ ] 默认不修改来源笔记。
- [ ] 写回时使用 create/patch 和 revision。
- [ ] 来源内容变化能被识别。

## 验证

```bash
./scripts/lint-skills.sh
go test ./... -run 'Skill.*(Compare|Synthesis)'
```

