# P2-T04：Daily Note 与模板一致性

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`miniobsidian.nvim`
- 依赖：P0-T02、P0-T05、P2-T01、P2-T02

## 目标

让插件创建的 Daily Note 与 Obsidian 官方 Daily Notes 配置及共同规范一致。

## 实施步骤

1. 扩展配置同步，读取 Daily Notes 的 folder、format 和 template。
2. 已配置模板时读取模板并执行规范允许的变量替换。
3. 未配置模板时使用可配置的最小默认内容。
4. 使用日历日期计算 yesterday/tomorrow，移除固定 `±86400`。
5. 模板选择使用相对路径，禁止 basename 静默覆盖。
6. 已存在 Daily Note 只打开，不覆盖。
7. 用共同 fixture 比较 CLI 与插件生成的目标路径和内容。

## 交付物

- Daily 配置同步
- 模板解析修复
- 语义一致性测试

## 验收标准

- [ ] Obsidian、CLI、插件计算出相同 Daily Note 路径。
- [ ] 配置模板被正确使用。
- [ ] 同名模板可消歧。
- [ ] 夏令时边界日期计算正确。
- [ ] 已存在文件内容不被覆盖。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/daily tests/template" -c qa
```

