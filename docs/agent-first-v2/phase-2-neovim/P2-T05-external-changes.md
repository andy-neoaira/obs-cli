# P2-T05：外部修改检测与缓存失效

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`miniobsidian.nvim`
- 依赖：P0-T04、P2-T01

## 目标

正确感知 Obsidian、同步工具和 Agent 对文件的外部修改，保护 Neovim 中未保存内容。

## 实施步骤

1. 在 `FocusGained`、`BufEnter` 等合理事件上执行受控 `checktime`。
2. 未修改 buffer 可提示并按配置自动重载。
3. 已修改 buffer 只提示冲突，禁止自动重载。
4. 提供查看 diff、保留 buffer、重新加载三个动作。
5. 外部创建、删除、重命名触发 note cache 失效。
6. 插件自身写入后立即精确失效，不依赖 5 秒 TTL。
7. 防止 autocmd 重入和大型 Vault 高频全量扫描。
8. 编写模拟外部写入的 headless 测试。

## 交付物

- 外部变更 autocmd
- 冲突提示与动作
- 缓存失效机制

## 验收标准

- [x] 未保存 buffer 不会被外部内容静默覆盖。
- [x] 外部新建笔记可及时进入补全和 picker。
- [x] 外部删除/重命名不会长期保留陈旧缓存。
- [x] 大型 Vault 不会在每次按键全量扫描。
- [x] 冲突路径有自动化测试。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
make test PLENARY_DIR=/path/to/plenary.nvim
make test NVIM=/path/to/nvim-0.10.4 PLENARY_DIR=/path/to/plenary.nvim
```

## 完成记录

- 2026-07-25：新增 `external_changes.lua`，在 `FocusGained` / `BufEnter` 上执行防抖的原生 `checktime`，并接管 Vault Markdown 的 `FileChangedShell`。
- 2026-07-25：默认保留内存 buffer 并提供统一 diff、保留 buffer、重新加载磁盘三个动作；新增 `:ObsidianResolveConflict`。
- 2026-07-25：`external_change_mode` 支持 `prompt`、`reload`、`notify`；`reload` 只对未修改 buffer 自动生效，未保存内容始终进入冲突流程。
- 2026-07-25：冲突存在时 `BufWritePre` 阻止普通和强制写入，避免陈旧 buffer 覆盖 Obsidian、同步工具或 Agent 的磁盘版本。
- 2026-07-25：新增防抖的 Vault `fs_event` 监听；插件自身 `BufWritePost` 立即失效缓存，外部创建、重命名、删除触发 O(1) 失效，FocusGained 与 5 秒 TTL 提供跨平台兜底。
- 2026-07-25：监听与 checktime 只处理事件或已加载 buffer，不在按键路径执行全 Vault 扫描，并具有重入与频率保护。
- 2026-07-25：新增 headless 自动化测试覆盖未保存冲突、默认提示、显式重载、diff、陈旧写入阻断及外部创建/重命名/删除。
- 2026-07-25：Stylua、Selene、fixture 校验、当前 Neovim 全套测试和 Neovim 0.10.4 全套测试全部通过。
