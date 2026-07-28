# P5 审计修复关闭报告

- 报告日期：2026-07-28
- 状态：`通过`
- 范围：`obs-cli` P5-T01 至 P5-T07
- 审计来源：[review-audit.md](../agent-first-v2/review-audit.md)

## 修复结果

| 任务 | 结果 | 提交 | 主要证据 |
|---|---|---|---|
| P5-T01 配置原子持久化 | 已修复 | `8e6e29a` | 配置复用跨平台 atomic replace，rename 后目录 sync，失败注入与 race 通过 |
| P5-T02 隐藏文件隔离 | 已修复 | `7ef0cf7` | 隐藏文件/目录不返回且不阻断普通列表，显式隐藏读取仍拒绝 |
| P5-T03 配置错误可见性 | 已修复 | `0e01738` | Daily 配置缺失与损坏分离，损坏配置返回 `INVALID_ARGUMENT` 且零写入 |
| P5-T04 V1 死代码 | 已修复 | `76ecee8` | 删除 16 个命令/辅助文件和 3 个专属测试，V2 根命令与迁移文档保留 |
| P5-T05 Capability 文档门禁 | 已修复 | `ab34300` | runtime flags 与规范表双向集合校验 |
| P5-T06 Skill 身份门禁 | 已修复 | `afc59eb` | 目录、frontmatter、eval manifest 三方身份一致 |

P5 任务基线由 `3e28eac` 固化。每项实现均作为独立提交，可单独审计和回滚。

## 审计结论修正

- M2 的真实问题是隐藏 Markdown 导致整个 List 失败，不是隐藏内容成功泄露；修复和
  测试采用真实行为。
- V2 `note.move` 不读取 `app.json`，因此 P5-T03 未增加虚假的配置依赖。
- `note.move --dry-run` 已符合统一顶层 dry-run 模型，未增加破坏兼容性的
  `result` 包装。
- Vault discovery 保持只读呈现不可用条目；规范化和符号链接校验继续在注册/迁移
  边界执行。
- Create/Move 的父目录创建前后已有路径身份复核，未当作越界漏洞修改。
- Vault registry dry-run 的 target/details 已能标识计划对象，未增加语义含糊的
  `vault` 字段。

## 联合验证

### obs-cli

- `make release-check`：通过。
- 全量 test 与 race：通过。
- 总覆盖率：`73.8%`，门槛 `70.0%`。
- Schema、六目标交叉构建、license、Skill lint/eval、RC smoke：通过。

### 三入口自动 E2E

- `./scripts/run-three-client-e2e.sh`：`6/6`。
- Obsidian 风格文件事件、Agent CLI、miniobsidian Adapter 的搜索、条件写入、
  冲突、移动、Daily 和 dirty buffer 场景全部通过。

### miniobsidian.nvim

- StyLua：通过。
- Selene：0 error、0 warning。
- 固定 Vault contract fixture：通过。
- 完整 Plenary suite：通过。
- 工作区未产生修改。

## 保留门禁

P5 完成不改变 P4 的发布结论。真实移动端同步延迟、离线冲突和冲突副本观察仍待
P4-T05 完成，因此：

- `docs/compatibility.json` 继续保持 `release-candidate`；
- P4-T05、P4-T06 和 P4 阶段继续保持未完成；
- 不创建或推送正式 tag/release。
