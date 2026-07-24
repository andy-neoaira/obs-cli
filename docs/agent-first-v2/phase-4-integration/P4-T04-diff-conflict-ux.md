# P4-T04：Agent 修改后的 Diff 与冲突 UX

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`miniobsidian.nvim`、`obs-cli`
- 依赖：P4-T03、P2-T05

## 目标

让用户在 Neovim 中清晰查看 Agent 的实际修改，并安全处理外部更新冲突。

## 实施步骤

1. Agent 响应返回 revision before/after、changed files 和摘要。
2. 插件收到结果后检查当前 buffer modified 状态。
3. 未修改 buffer 可执行 `checktime` 并展示 diff。
4. 已修改 buffer 打开三方比较：原 revision、Agent 结果、本地 buffer。
5. 提供保留本地、采用磁盘、手动合并动作。
6. 禁止自动选择冲突解决方案。
7. 多文件结果提供 changed-files picker。
8. 记录失败、取消和部分失败的用户可操作说明。

## 交付物

- Agent result handler
- Diff/三方比较 UI
- 冲突场景测试

## 验收标准

- [ ] 用户能看到 Agent 实际修改了哪些文件。
- [ ] dirty buffer 永不自动重载。
- [ ] 冲突处理动作不会丢失任一版本。
- [ ] `PARTIAL_FAILURE` 能展示恢复清单。
- [ ] 取消操作不继续写入。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/integration/conflict" -c qa
```

