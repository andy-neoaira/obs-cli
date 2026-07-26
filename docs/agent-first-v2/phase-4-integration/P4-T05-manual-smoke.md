# P4-T05 人工 Obsidian 冒烟清单

> 本清单必须使用专用测试 Vault。禁止使用个人 Vault。自动化脚本不能替代
> Obsidian 桌面端和移动端的真实行为验证。

## 测试信息

- 状态：`桌面端通过，移动端待人工执行`
- 测试人：`Codex（桌面端）/ 待分配（移动端）`
- 桌面端版本：`Obsidian 1.12.7`
- 移动端平台与版本：`待填写`
- 同步方式与版本：`桌面端使用本地专用 Vault；移动端待填写`
- 执行日期：`2026-07-25`
- 测试 Vault：`/private/tmp/obs-cli-desktop-smoke.BrHdLW/vault（专用临时 Vault）`

## 前置门禁

- [x] `./scripts/run-three-client-e2e.sh` 输出 `6/6`。
- [x] 测试 Vault 已备份或可随时丢弃。
- [ ] 桌面端、移动端、CLI 和 Neovim 指向同一份同步数据。
- [ ] 已记录同步服务的冲突副本、延迟与离线策略。

## 桌面端

- [x] 在 Obsidian 创建含 Frontmatter 与 Wikilink 的笔记。
- [x] Agent 通过 CLI 搜索并以 revision 前置条件更新该笔记。
- [x] Neovim 打开后内容、Frontmatter 和链接一致。
- [x] Neovim 新建笔记后，Obsidian 文件列表和图谱可发现它。
- [x] CLI 移动笔记并重写链接后，Obsidian 不出现新增坏链接。
- [x] Obsidian 外部修改发生在 Agent plan/apply 之间时，apply 返回冲突且不覆盖。
- [x] 三个入口打开同一天的 Daily Note，路径和模板结果一致。

## 移动端与同步

- [ ] 移动端能打开桌面端/Agent/Neovim 创建的笔记。
- [ ] 移动端编辑同步回桌面后，CLI 读取到新的 revision。
- [ ] 离线编辑与 Agent 并发更新时，不发生无提示的数据覆盖。
- [ ] 若同步服务生成冲突副本，已记录命名、位置和恢复步骤。
- [ ] 已测量并记录典型同步延迟；测试期间不把“尚未同步”误判为文件丢失。
- [ ] 已确认移动端插件限制不会改变 Vault 文件格式或 Daily Note 路径。

## 结果记录

| 项目 | 结果 | 证据/备注 |
|---|---|---|
| 桌面端闭环 | 通过 | Obsidian 创建 `Desktop Smoke Created.md`；CLI 搜索并条件更新；Neovim 读取 Frontmatter、正文和 Wikilink 一致。 |
| 移动端读取与回写 | 待测 | |
| 并发冲突保护 | 通过 | Obsidian 在 plan/apply 间更新 `Conflict.md` 后，CLI 以退出码 4 返回 `REVISION_CONFLICT`，磁盘保留桌面端修改。 |
| Daily Note 一致性 | 桌面三入口通过 | Obsidian、CLI 与 Neovim 均解析为 `Dailies/2026-07-27.md`；移动端待测。运行主机日期与报告日期存在偏差，因此保留实际解析路径作为证据。 |
| 移动与链接完整性 | 通过 | CLI 将笔记移动并重命名为 `Archive/Desktop Smoke Renamed.md`，`Refs.md` 自动改写为 `[[Desktop Smoke Renamed]]`，backlinks 返回 1 条且旧链接不存在。 |
| 同步限制 | 待记录 | |

## 桌面端证据

- CLI Vault：`DesktopSmoke`（`vlt_c4d8424daad716b992bad590305a481d`）。
- CLI 条件写入使用真实 revision；Obsidian 编辑区自动刷新为
  `Agent verified desktop marker.`。
- `miniobsidian.nvim` 真实 headless Adapter 成功读取桌面端创建的标题、标签、
  正文和 `[[Refs]]`，并创建 `Notes/desktop-from-neovim.md`；Obsidian 文件列表
  自动发现该笔记。
- 第二次移动使用 dry-run 返回的 revision 与 plan hash 执行；Obsidian 自动刷新
  `Refs.md`，未观察到新增坏链接。
- stale apply 返回 `REVISION_CONFLICT` 后，`Conflict.md` 仍为
  `Obsidian desktop external edit wins.`，不存在 Agent 计划中的替换文本。
- 测试全过程只使用可丢弃的临时 Vault；结束后已恢复
  `AndyObsidian` 原始窗口，未修改个人 Vault 内容。

只有所有必测项通过，且限制与恢复步骤已记录后，才能把 P4-T05 标记为
`已完成`。
