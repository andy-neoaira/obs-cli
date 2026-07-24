# ADR-001：Agent-first 产品边界与三入口架构

- 状态：已接受
- 决策日期：2026-07-24
- 决策范围：`obs-cli` V2、场景化 Skills、`miniobsidian.nvim`、Obsidian Vault
- 关联计划：[Agent-first V2 联合改造执行计划](../agent-first-v2/README.md)

## 背景

同一份 Obsidian 笔记需要由三类入口持续读写：

1. Obsidian 桌面端和移动端，负责通用的人工作业与同步体验。
2. AI Agent，通过 `obs-cli` 和场景化 Skills 完成阅读、检索、比较、分析与安全更新。
3. Neovim，通过 `miniobsidian.nvim` 完成 Vimer 的原生编辑、导航和快速记录。

当前 `obs-cli` 源自面向人类终端操作的 `notesmd-cli`，其交互模式、输出形式和隐式副作用不适合作为可靠的 Agent 执行协议。`miniobsidian.nvim` 已经能够独立完成核心 Neovim 工作流，如果改为强制调用 CLI，会增加安装门槛、进程调用开销和跨项目版本耦合。

必须在 V2 实现前冻结产品边界，避免两个项目重复实现复杂批处理逻辑，或形成 `Neovim → CLI → Obsidian` 的错误依赖链。

## 决策

### 1. Vault 是唯一内容事实源

Markdown、附件和 Vault 内配置组成唯一内容事实源。Obsidian、`obs-cli` 和 `miniobsidian.nvim` 都是操作该事实源的同级客户端，不以任一应用的进程或私有状态作为内容主库。

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

同步服务可以搬运文件，但不改变上述事实源和客户端边界。

### 2. obs-cli V2 是 Agent-first 执行器

`obs-cli` V2 的首要消费者是 AI Agent、Skill、脚本和自动化系统。核心能力必须：

- 非交互、可组合并具有稳定 JSON 协议。
- 显式声明 Vault、目标、写入意图和前置条件。
- 支持 capability discovery、dry-run、revision、原子写入和稳定错误码。
- 默认拒绝路径逃逸、目标覆盖、版本冲突和歧义目标。
- 将 stdout 用于机器数据，将 stderr 用于诊断信息。

人类可以直接调用 CLI，但 Human-first 体验不得改变核心协议语义。

### 3. Skills 是场景层

Skills 负责触发语义、上下文读取、计划、授权、冲突处理和结果验证；CLI 负责可预测的原子操作。

业务场景不直接固化为不透明的大型 CLI 命令。写入型 Skill 必须遵循：

```text
discover → read with revision → plan/dry-run → authorize → apply → verify
```

### 4. miniobsidian.nvim 是独立 Neovim 客户端

`miniobsidian.nvim` 不强制依赖 `obs-cli`，其基础功能在没有 CLI 的环境中必须完整可用。笔记编辑、buffer 管理、补全、picker、checkbox、模板插入和图片粘贴继续由插件原生实现。

插件未来可以通过可选 adapter 使用 CLI 的高级能力，例如安全移动并重写链接、Vault 审计和 Agent 任务交接。可选 adapter 必须通过 capability 协商启用，失败时不得破坏插件的基础能力。

反向依赖同样禁止：`obs-cli` 不依赖 Neovim、插件状态或 Neovim RPC。

### 5. 共享规范，不共享运行时依赖

两个项目共同遵守版本化的 Vault 约定和测试 fixture，覆盖：

- 路径规范化与 Vault 边界。
- Frontmatter 和 Note ID。
- Wikilink 解析与同名消歧。
- Daily Note 目录、日期格式和模板。
- 附件与相对链接。
- revision、冲突检测和写入安全。

Go 与 Lua 可以独立实现，但相同 fixture 必须产生一致的语义结果。

### 6. 配置所有权

- `.obsidian/` 和 Obsidian 官方全局配置由 Obsidian 所有。
- `obs-cli` 可以只读发现或显式导入 Obsidian 配置，不直接写入官方全局配置。
- `miniobsidian.nvim` 可以只读同步所需的 Vault 设置，不把插件私有配置写入 `.obsidian/`。
- 两个工具的私有配置分别存储，私有设置不能隐式改变另一个客户端。
- 共同内容约定通过版本化规范表达，不通过共享可变配置文件暗中耦合。

## 各组件职责

| 组件 | 负责 | 不负责 |
|---|---|---|
| Obsidian | 桌面/移动编辑、官方插件体验、官方配置、同步入口 | 为 CLI 或 Neovim 提供必须在线的服务 |
| obs-cli V2 | 机器协议、安全原子操作、批处理计划、冲突与错误模型 | TTY picker、内置编辑器体验、Neovim UI |
| 场景化 Skills | 意图理解、上下文编排、授权、验证和恢复策略 | 绕过 CLI 安全规则直接改 Vault |
| miniobsidian.nvim | Neovim buffer、导航、补全、原生编辑体验 | Agent 业务编排、复杂多文件事务的重复实现 |
| 共同规范与 fixture | 统一文件语义和跨实现验收 | 成为运行时服务或强制共享代码库 |

## V2 非目标

以下能力不属于 Agent-first 核心：

- 依赖 TTY 的交互式确认或问答。
- CLI 内置 fuzzy picker。
- CLI 自动启动 `$EDITOR` 或其他编辑器。
- CLI 隐式打开 Obsidian、浏览器或 Obsidian URI。
- 根据当前工作目录、焦点窗口或上次交互隐式推断写入目标。
- 为保持 V1 表面兼容而保留不安全或语义含糊的行为。
- 让 CLI 承担 Obsidian 同步服务或数据库职责。
- 让 `miniobsidian.nvim` 成为 CLI 的图形前端。

可选的人类便利功能只有在不影响核心协议、可被完全禁用且不成为 Skill 依赖时，才可以作为外围适配层存在。

## 备选方案

### 方案 A：miniobsidian.nvim 强制依赖 obs-cli

未采用。它会提高纯 Neovim 用户的安装门槛，引入外部进程延迟和版本耦合，而且 buffer、补全、picker 等能力无法从 CLI 获益。

### 方案 B：obs-cli 与插件完全独立演化

未采用。虽然运行时解耦，但 Daily Note、路径、链接和冲突语义会继续漂移，最终让三个入口对同一文件产生不同理解。

### 方案 C：重新创建 obs-cli 项目

未采用。现有 Go 项目的 Vault、Frontmatter、链接和测试基础可以复用。采用当前仓库内的破坏性 V2 重构，风险和迁移成本更低。

### 方案 D：建立常驻服务或中心数据库

当前不采用。它会破坏本地 Markdown 优先和离线可用的模型，也会让手机端与原生 Obsidian 产生额外集成负担。未来只有在文件级并发协议被证明不足时才重新评估。

## 后果

### 正向后果

- 三个入口都能独立工作，任一工具离线不会锁住笔记。
- Agent 写入具备可验证、可冲突检测和可恢复的协议基础。
- 插件保持轻量，对只使用 Neovim 的用户没有额外安装要求。
- 复杂多文件操作集中在 CLI，避免 Lua 重复实现事务逻辑。
- 共同 fixture 能及时发现跨入口语义漂移。

### 成本与约束

- Go 与 Lua 需要分别实现部分相同语义，并共同维护测试向量。
- `obs-cli` V2 会发生不兼容命令和配置变更。
- revision 使用乐观并发控制，冲突需要 Agent 或用户重新合并。
- 文件同步的延迟和冲突行为仍受具体同步服务影响。
- 可选 CLI adapter 需要维护协议兼容矩阵。

## 实施约束

1. 后续 P0 任务必须把本决策细化为版本化规范和 fixture。
2. 任何新增跨项目依赖都需要新的 ADR。
3. 任何写入命令不得绕过统一路径和原子存储层。
4. 任何插件高级 CLI 集成都必须是可选且可安全降级的。
5. 修改本决策时，应同步更新两个项目 README 和联合任务台账。

## 验收映射

| 决策要求 | 后续验收任务 |
|---|---|
| Vault 共同语义 | P0-T02、P0-T05 |
| JSON 与 capability | P0-T03、P1-T04、P1-T05 |
| revision 与原子写入 | P0-T04、P1-T03 |
| 插件独立可靠 | P2、P4-T01 |
| 场景化 Agent 工作流 | P3 |
| 三入口闭环 | P4-T05 |

