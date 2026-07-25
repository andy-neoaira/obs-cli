# Three-client E2E fixture

该 fixture 只包含合成笔记，用于模拟 Obsidian 文件语义、`obs-cli` Agent
操作和 `miniobsidian.nvim` 操作同一 Vault。执行脚本会把 `seed/` 复制到
临时目录，绝不在此目录或个人 Vault 上直接写入。

`golden-summary.json` 只记录稳定的场景与不变量，不包含临时路径、时间戳、
Vault ID 或 revision，因此可跨机器比较。
