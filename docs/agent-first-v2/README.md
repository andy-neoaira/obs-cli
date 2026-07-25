# Agent-first V2 联合改造执行计划

本目录是 `obs-cli`、`miniobsidian.nvim` 与 Obsidian 三入口协同改造的任务事实源。后续开发以这里的任务编号、依赖关系和验收标准为准。

## 目标架构

```text
                  同一个 Obsidian Vault
                         │
       ┌─────────────────┼─────────────────┐
       │                 │                 │
 Obsidian App       obs-cli V2      miniobsidian.nvim
 手机 / 桌面端       Agent 执行器        Neovim 客户端
                         │
                   场景化 Skills
```

- Markdown Vault 是唯一内容事实源。
- `obs-cli` 是 Agent-first、安全、可组合的 Vault 操作内核。
- `miniobsidian.nvim` 保持独立，不强制依赖 `obs-cli`。
- 三个入口共享语义规范，不共享运行时依赖。
- 所有 Agent 写入均支持 revision 前置条件、冲突检测和原子落盘。

## 仓库范围

| 项目 | 本地路径 | 职责 |
|---|---|---|
| obs-cli | `/Users/andy/github/obs-cli` | Agent CLI、Skills、联合规范与任务台账 |
| miniobsidian.nvim | `/Users/andy/github/miniobsidian.nvim` | Neovim 原生交互入口 |
| Obsidian | 用户 Vault 与 `.obsidian/` | 桌面端、移动端及官方配置 |

## 状态约定

- `[ ]`：未开始
- `[-]`：进行中
- `[x]`：已完成并通过验收
- `[!]`：阻塞，必须在任务文档中记录原因

更新任务状态时，必须同时更新：

1. 对应子任务文档中的“状态”。
2. 所属阶段 `README.md` 的任务复选框。
3. 本文件中的阶段状态。

## 阶段总览

| 阶段 | 内容 | 状态 | 完成条件 |
|---|---|---|---|
| [P0 共同协议与基线](./phase-0-contract/README.md) | 产品边界、Vault 规范、协议、并发模型、测试夹具、许可证 | `[x]` | 两个实现拥有可测试的共同语义 |
| [P1 obs-cli 安全内核](./phase-1-cli-core/README.md) | Agent-first CLI、原子写入、JSON、revision、事务化操作 | `[x]` | CLI 可安全支撑写入型 Agent |
| [P2 Neovim 可靠性](./phase-2-neovim/README.md) | 测试、路径、链接、Daily、外部变更、UX 一致性 | `[x]` | 插件独立可靠并与共同规范一致 |
| [P3 场景化 Skills](./phase-3-skills/README.md) | Skill 契约、迁移、新场景、评测 | `[x]` | Agent 能按场景安全完成闭环任务 |
| [P4 可选协同与端到端验收](./phase-4-integration/README.md) | 可选 CLI 适配、Agent 交接、冲突 UX、三入口 E2E | `[ ]` | 三入口操作同一 Vault 且不会静默覆盖 |

## 执行规则

1. 默认按 P0 → P1/P2 → P3 → P4 执行。
2. P1 与 P2 可在 P0 完成后并行开发。
3. 一个子任务应对应一个独立 PR 或一组可独立回滚的提交。
4. 不因任务实施而扩大范围；新需求先新增任务文档并标注依赖。
5. 修改共同语义时，必须同步更新规范、共享 fixture 和两个实现的测试。
6. 任务只有在交付物齐全、验收项全部勾选、验证命令通过后才能标记完成。

## 里程碑

- **M1：协议冻结**：P0 完成。
- **M2：安全写入可用**：P1-T01 至 P1-T06 完成。
- **M3：双客户端语义一致**：P1、P2 完成。
- **M4：Agent 场景可发布**：P3 完成。
- **M5：信息闭环完成**：P4 完成。
