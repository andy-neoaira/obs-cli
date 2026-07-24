# P4-T03：当前笔记与选区交给 Agent

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`miniobsidian.nvim`、Agent 集成
- 依赖：P4-T01、P3-T06、P3-T07

## 目标

允许用户从 Neovim 将当前笔记、选区和任务意图安全地交给 Agent 分析或更新。

## 实施步骤

1. 定义 handoff payload：Vault ID、相对路径、revision、选区、意图。
2. 默认仅传路径和选区，不隐式传整个 Vault。
3. 增加 `:ObsidianAgentAnalyze` 和 `:ObsidianAgentUpdate`。
4. buffer 有未保存修改时，要求保存或仅发送内存内容做只读分析。
5. 更新型 handoff 必须指定允许写入的目标路径。
6. Agent 使用对应 Skill 和 capability 检查。
7. 保存 request ID，便于关联执行结果。
8. 对敏感内容和大选区增加预览/确认配置。

## 交付物

- handoff payload 规范
- Neovim 命令与 Agent 接口
- 权限边界测试

## 验收标准

- [ ] Agent 获得明确路径和 revision。
- [ ] 默认不会读取未授权的整个 Vault。
- [ ] 未保存 buffer 的行为明确且无数据丢失。
- [ ] 更新请求能追踪到 request ID。
- [ ] 只读分析不会触发文件修改。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/integration/handoff" -c qa
```

