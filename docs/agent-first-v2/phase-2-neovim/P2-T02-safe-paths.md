# P2-T02：Vault 路径安全模块

- 状态：`已完成`
- 负责人：`Codex`
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

- [x] 不再用裸字符串前缀判断 Vault 边界。
- [x] `../` 和符号链接逃逸被拒绝。
- [x] 合法 Unicode、空格、相对路径正常工作。
- [x] 附件和模板遵守同一规则。
- [x] Windows 路径行为有测试或明确的平台限制。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
rg -n 'sub\\(1,.*#vault|starts_with' lua/miniobsidian
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/path" -c qa
```

## 完成记录

- 完成日期：2026-07-25
- 实现提交：`miniobsidian.nvim@ac34734`
- 新增统一 `miniobsidian.path` 模块，提供 normalize、realpath、join、relative、`is_within_vault` 与安全 resolve。
- 新目标通过最近存在父目录解析真实路径；现有目标和父目录中的符号链接均在操作前完成 Vault 边界校验。
- `note`、`daily`、`template`、`image`、`vault`、笔记扫描和 buffer 边界判断均已迁移，不再使用字符串前缀充当安全判断。
- 普通内容路径拒绝绝对路径、`~`、NUL、空段、`.`、`..`、隐藏目录、Windows ADS、保留设备名及尾随点/空格。
- Windows 分隔符输入统一为 `/`；Windows 运行时的边界比较采用大小写不敏感语义。UNC 和盘符绝对路径按共同协议拒绝。
- 路径测试覆盖 Unicode、空格、相似目录前缀、最近父目录、内部/外部符号链接，以及共享 Windows fixture。
- note、daily、template 和 attachments 配置的 `../` 逃逸均有操作级回归测试。
- `make ci` 通过：Stylua、Selene、共享 fixture 和 20 个测试全部成功。
- Neovim v0.10.4 最低支持版本补跑同一测试集，20 个测试全部成功。
