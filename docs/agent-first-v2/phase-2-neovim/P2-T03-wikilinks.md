# P2-T03：Wikilink 解析、消歧与跳转

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`miniobsidian.nvim`
- 依赖：P0-T02、P0-T05、P2-T01、P2-T02

## 目标

完整支持 `[[path|alias#heading^block]]` 语义，并禁止同名笔记静默误跳。

## 实施步骤

1. 将 Wikilink tokenizer/parser 与文件查找分离。
2. 保留并解析 path、alias、heading 和 block。
3. 优先按规范化相对路径匹配。
4. basename 多匹配时弹出消歧选择或返回明确错误。
5. 创建不存在笔记时保留用户指定目录。
6. 打开后定位 heading 或 block。
7. 补全项显示可消歧路径，插入稳定 link target。
8. 使用共享 fixture 测试 Unicode、alias、heading、block 和重复名称。

## 交付物

- Wikilink parser
- resolver 与定位逻辑
- 补全和跳转测试

## 验收标准

- [x] `[[folder/note]]` 不会跳到其他目录同名笔记。
- [x] 多匹配不会自动选择第一个。
- [x] heading 和 block 可正确定位。
- [x] alias 不影响目标解析。
- [x] 创建路径与 link 中的目录一致。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/link" -c qa
```

## 完成记录

- 完成日期：2026-07-25
- 实现提交：`miniobsidian.nvim@d2b2f1f`
- 新增独立 `miniobsidian.wikilink` parser/resolver，分别保留 target、alias、heading 与 block ID。
- 含目录 target 只按 Vault 相对 Note ID 精确匹配；basename 按精确大小写优先、唯一大小写折叠匹配兜底。
- basename 多匹配返回 `AMBIGUOUS_NOTE` 与稳定排序候选，并通过 `vim.ui.select` 要求用户显式消歧。
- 不存在的限定路径在确认创建时保留原目录和文件名，不再退化到默认笔记目录。
- 跳转后支持 Markdown heading、重复 heading anchor 和精确 block ID 定位；fragment 不存在时仍打开笔记并给出独立提示。
- 补全对同名笔记显示并插入带目录的稳定 Note ID，唯一 basename 继续使用短链接。
- 测试覆盖 alias、heading、block、大小写、同名消歧、目录创建、稳定补全及共享 duplicate-note fixture。
- `make ci` 通过：Stylua、Selene、共享 fixture 和 27 个测试全部成功。
- Neovim v0.10.4 最低支持版本补跑同一测试集，27 个测试全部成功。
