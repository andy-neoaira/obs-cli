# P0-T06：开源许可与发布基线

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli`、`miniobsidian.nvim`
- 依赖：P0-T01

## 目标

明确派生关系、第三方许可证和发布物包含内容，降低 V2 发布后的合规风险。

## 实施步骤

1. 在 `obs-cli` README 标明派生自 `Yakitrak/notesmd-cli`。
2. 保留上游 MIT copyright，并追加本项目新增代码声明。
3. 生成 `THIRD_PARTY_NOTICES.md`，列出直接和 vendored 依赖许可证。
4. 检查 GoReleaser 配置，确保源码包和二进制包包含必要许可证。
5. 检查 `miniobsidian.nvim` 是否存在复制代码；如有，保留对应声明。
6. 在插件 README 保留灵感来源说明。
7. 在发布检查中增加许可证文件存在性验证。

## 交付物

- `obs-cli/THIRD_PARTY_NOTICES.md`
- 更新后的两个项目 README
- 发布物许可证校验脚本或 CI 步骤

## 验收标准

- [x] 未移除上游 MIT 声明。
- [x] 派生来源在 README 可见。
- [x] 二进制归档配置包含项目 LICENSE、第三方声明和依赖许可原文。
- [x] 插件自身 MIT 声明与当前可审查代码来源一致。
- [x] 发布检查缺少许可证时失败。

## 验证

```bash
rg -n "Yakitrak/notesmd-cli|MIT|THIRD_PARTY" README.md LICENSE THIRD_PARTY_NOTICES.md
goreleaser release --snapshot --clean
```

## 验证记录

- 2026-07-24：`make license-check` 通过，23 个 vendored module 均有许可文件和 notices 条目。
- 2026-07-24：保留上游 copyright，并追加本项目修改部分声明。
- 2026-07-24：GoReleaser `archives.files` 显式包含 `LICENSE`、`THIRD_PARTY_NOTICES.md`、依赖 LICENSE/COPYING/NOTICE。
- 2026-07-24：CI 与 release workflow 均接入 `make license-check`。
- 2026-07-24：插件保留独立 MIT License 和 `obsidian.nvim` 灵感来源说明；审查边界记录在 `docs/legal/LICENSE_REVIEW.md`。
- 说明：本机未安装 GoReleaser，归档内容由配置、许可门禁和 GitHub Release workflow 校验；首次 V2 RC 仍须按 P1-T08 执行 snapshot 归档实测。
