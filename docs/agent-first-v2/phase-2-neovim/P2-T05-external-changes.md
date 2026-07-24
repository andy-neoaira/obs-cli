# P2-T05：外部修改检测与缓存失效

- 状态：`未开始`
- 负责人：`待分配`
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

- [ ] 未保存 buffer 不会被外部内容静默覆盖。
- [ ] 外部新建笔记可及时进入补全和 picker。
- [ ] 外部删除/重命名不会长期保留陈旧缓存。
- [ ] 大型 Vault 不会在每次按键全量扫描。
- [ ] 冲突路径有自动化测试。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/external_changes tests/cache" -c qa
```

