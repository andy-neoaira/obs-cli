# P0-T04：Revision、原子写入与冲突协议

- 状态：`已完成`
- 负责人：`Codex`
- 涉及项目：`obs-cli`、`miniobsidian.nvim`
- 依赖：P0-T02、P0-T03

## 目标

定义跨 Obsidian、Agent 和 Neovim 的乐观并发控制，禁止基于旧内容静默覆盖新修改。

## 实施步骤

1. 在共同规范中定义 revision 算法，建议使用文件原始字节的 SHA-256。
2. 定义读取操作必须返回 revision。
3. 定义 `--if-match` 的校验时机和 `REVISION_CONFLICT`。
4. 定义 create、append、patch、replace、move、delete 的并发语义。
5. 定义临时文件、flush、close、原子 rename 的写入顺序。
6. 定义目标已存在、跨文件系统 rename 和中断恢复行为。
7. 定义 Neovim 存在未保存 buffer 时的提示策略。
8. 编写手机同步后 Agent 旧 revision 写入的时序示例。

## 交付物

- `docs/spec/CONCURRENCY_AND_WRITES.md`
- revision 计算测试向量
- 冲突时序示例

## 验收标准

- [x] 同一字节内容在 Go 与 Lua 验证中产生相同 revision。
- [x] 协议明确 revision 不匹配时禁止修改目标文件。
- [x] 协议明确成功、冲突和失败后的临时文件清理规则。
- [x] 协议定义故障注入矩阵，要求进程中断不能产生半写入 Markdown。
- [x] 协议明确未保存 Neovim buffer 不得被自动重载覆盖。

## 验证

```bash
GOCACHE=/private/tmp/obs-cli-gocache \
  go test ./pkg/obsidian -run RevisionContractVectors

NVIM_LOG_FILE=/private/tmp/obs-vault-contract-nvim.log \
  nvim --headless -u NONE -i NONE \
  "+lua <读取 revision-v1.json 并用 vim.fn.sha256 校验>" +qa
```

## 验证记录

- 2026-07-24：5 组原始字节向量通过 Go SHA-256 测试。
- 2026-07-24：同一组向量通过 Neovim Lua `vim.fn.sha256` 校验。
- 2026-07-24：规范覆盖 create/append/patch/replace/move/delete 的前置条件。
- 2026-07-24：规范覆盖单文件原子提交、多文件回滚、partial failure、dirty buffer 与故障注入。
- 说明：实际存储层行为由 P1-T03 实现和验收，避免 P0/P1 循环依赖。
