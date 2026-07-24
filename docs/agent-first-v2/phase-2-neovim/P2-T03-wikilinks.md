# P2-T03：Wikilink 解析、消歧与跳转

- 状态：`未开始`
- 负责人：`待分配`
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

- [ ] `[[folder/note]]` 不会跳到其他目录同名笔记。
- [ ] 多匹配不会自动选择第一个。
- [ ] heading 和 block 可正确定位。
- [ ] alias 不影响目标解析。
- [ ] 创建路径与 link 中的目录一致。

## 验证

```bash
cd /Users/andy/github/miniobsidian.nvim
nvim --headless -u tests/minimal_init.lua -c "PlenaryBustedDirectory tests/link" -c qa
```

