# 开源许可与来源审查

- 审查日期：2026-07-24
- 范围：`obs-cli`、`miniobsidian.nvim`
- 性质：工程合规检查，不构成法律意见

## obs-cli

`obs-cli` 派生自 `Yakitrak/notesmd-cli`。上游采用 MIT License，当前 `LICENSE` 保留 Kartikay Jainwal 的原始版权与完整 MIT 许可文本，并追加本项目修改部分的版权声明。

Vendored Go 模块、版本、许可证类型和原始许可文件路径记录在根目录 `THIRD_PARTY_NOTICES.md`。发布归档显式包含项目 LICENSE、第三方声明，以及 vendor 中的 LICENSE、COPYING 和 NOTICE 文件。

`scripts/license-check.sh` 检查：

- 上游和当前项目版权声明。
- README 派生来源说明。
- 每个 vendored module 是否有许可文件。
- 每个 vendored module 是否出现在第三方声明中。
- GoReleaser 是否包含必要许可文件。

## miniobsidian.nvim

插件使用独立 MIT License，版权为 `andy-neoaira`。本次审查覆盖当前 Git 历史、源码文件和 README，没有发现文件级第三方版权头或明确复制来源；README 已保留对 `obsidian.nvim` 的设计灵感说明。

该结论不能证明不存在未记录的片段级复制。如果未来引入或复制第三方实现，提交者必须在合并前确认原许可证兼容性，并保留要求的版权、LICENSE 和 NOTICE。

