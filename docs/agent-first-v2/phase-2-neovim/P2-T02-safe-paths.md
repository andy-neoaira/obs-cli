# P2-T02：Vault 路径安全模块

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`miniobsidian.nvim`
- 依赖：P0-T02、P0-T05、P2-T01

## 目标

替换字符串前缀式 Vault 判断，为笔记、模板和附件建立统一路径安全模块。

## 实施步骤

1. 新建 `lua/miniobsidian/path.lua`。
2. 封装 normalize、realpath、join、relative 和 `is_within_vault`。
3. 对不存在目标解析最近存在父目录。
4. 处理路径分隔符、大小写和符号链接。
5. 将 `note`、`daily`、`template`、`image`、`vault` 接入该模块。
6. 验证 `attachments_folder` 不能通过 `..` 逃逸。
7. 使用共享 fixture 编写跨平台边界测试。

## 交付物

- 统一 Lua 路径模块
- 调用方迁移
- 路径安全测试

## 验收标准

- [ ] 不再用裸字符串前缀判断 Vault 边界。
- [ ] `../` 和符号链接逃逸被拒绝。
- [ ] 合法 Unicode、空格、相对路径正常工作。
- [ ] 附件和模板遵守同一规则。
- [ ] Windows 路径行为有测试或明确的平台限制。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
rg -n 'sub\\(1,.*#vault|starts_with' lua/miniobsidian
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/path" -c qa
```

