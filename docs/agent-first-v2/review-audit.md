# obs-cli agent-first-v2 代码审计报告

> 审计范围：`docs/agent-first-v2` 下已标记完成的 P0-P4 任务在当前代码库中的实现。
> 约束：**只读审计，未修改任何代码。**

## 1. 审计方法

1. 阅读子代理给出的候选问题清单。
2. 回到具体源码位置核验行号、行为与任务规范/Spec 是否一致。
3. 对无法复现或已被测试/Schema 覆盖的项予以排除。
4. 将确认后的发现按 Major / Minor / Informational 分级记录。

## 2. 问题总览

| 等级 | 数量 | 说明 |
|---|---:|---|
| Major | 2 | 影响正确性或与核心规范冲突 |
| Minor | 8 | 一致性、可维护性或文档缺口 |
| Informational | 2 | 状态说明或计划内行为 |

## 3. Major 问题

### M1. 配置写入未复用原子存储层且缺少目录 fsync

- **位置**：`pkg/config/v2_store.go:182-216`
- **现象**：`Store.writeAtomic` 自行实现 `temp-create → chmod → write → sync → close → rename`，未调用 `pkg/storage` 的 `WriteAtomic`，也没有在 `os.Rename` 后执行目录 `fsync`。
- **不合理原因**：
  - P1-T03 明确要求“所有写命令不再直接调用 `os.WriteFile` 或裸 `os.Rename`”，配置写入同样是写操作。
  - 自行实现导致原子写入逻辑重复；缺少目录 `fsync` 在部分文件系统崩溃场景下可能导致“文件已写但目录项未持久化”。
- **推荐方案**：
  - 将配置写入委托给 `storage.Store.WriteAtomic`；若其 precondition 模型不支持无条件覆盖，可新增 `Preconditions{Overwrite: true}`。
  - 在过渡方案中，至少补齐 `syncDirectory(filepath.Dir(s.path))`，与 `pkg/storage/atomic.go:156` 保持一致。
- **状态**：已确认。

### M2. `noteops.List` 未过滤隐藏文件，与路径策略不一致

- **位置**：`pkg/noteops/service.go:94-128`
- **现象**：`filepath.WalkDir` 只跳过 `entry.IsDir() && strings.HasPrefix(entry.Name(), ".")` 的隐藏目录；隐藏 Markdown 文件（如 `.draft.md`）仍会被列出。
- **不合理原因**：
  - `pkg/pathpolicy` 默认拒绝隐藏路径段（规则 `VC-2.3`）。
  - `List` 列出隐藏笔记后，`note.get` 等后续操作会失败，造成“列表可见、读取被拒”的不一致，也可能泄露用户本意隐藏的笔记。
- **推荐方案**：在遍历中过滤任何路径段以 `.` 开头的条目；或统一使用 `pathpolicy.Resolver`（`AllowHidden=false`）判断可见性。
- **状态**：已确认。

## 4. Minor 问题

### m1. `noteops.Create` / `Move` 直接使用 `os.MkdirAll`

- **位置**：`pkg/noteops/service.go:195`；`pkg/noteops/move.go:184`
- **现象**：创建笔记或移动目标父目录时直接调用 `os.MkdirAll`，未经过路径策略校验与存储层抽象。
- **不合理原因**：
  - P1-T03 完成记录称“已移除业务层直接文件写入”。目录创建虽不直接写入 Note 内容，但仍可能创建隐藏目录或越界父目录，且与存储层原子语义割裂。
- **推荐方案**：在 `storage.Store` 增加 `EnsureParentDir(path)`，或在业务层先对父目录做 `resolver.Resolve` 校验，再由存储层统一创建。
- **状态**：已确认。

### m2. 未注册的 V1 遗留命令仍保留在代码树中

- **位置**：
  - V1 命令：`cmd/add_vault.go`、`cmd/create.go`、`cmd/daily.go`、`cmd/delete.go`、`cmd/editor.go`、`cmd/frontmatter.go`、`cmd/list.go`、`cmd/list_vaults.go`、`cmd/move.go`、`cmd/open.go`、`cmd/print.go`、`cmd/remove_vault.go`、`cmd/search.go`、`cmd/search_content.go`、`cmd/set_default.go`。
  - 仅被 V1 命令引用：`cmd/content_input.go`、`cmd/content_input_test.go`。
- **现象**：`cmd/root.go:42-53` 只注册 V2 命令，上述文件在运行时已不被使用（`grep` 未找到任何对 `newCreateCommand` 等构造函数的引用）。
- **不合理原因**：死代码增加编译体积、测试维护成本，并存在未来被误注册或误导用户的风险。
- **推荐方案**：在确认无外部引用后删除，或迁移到 `cmd/legacy/` 归档目录；同步清理 `content_input.go` 及其测试。
- **状态**：已确认。

### m3. `capabilities` 中 `move_plan_preconditions` feature flag 未在规范文档中解释

- **位置**：`cmd/capabilities.go:70`；`docs/spec/CAPABILITIES.md` 第 2 节 feature flag 表格。
- **现象**：运行时返回 `move_plan_preconditions: true`，但规范文档的表格未列出该 flag。
- **不合理原因**：Agent/Skill 调用方依赖规范文档理解 capability；未文档化的 flag 会降低可发现性，也可能导致调用方不敢依赖。
- **推荐方案**：在 `docs/spec/CAPABILITIES.md` 表格补充 `move_plan_preconditions` 的语义说明。
- **状态**：已确认。

### m4. `note move --dry-run` 响应结构与统一 dry-run 模型不一致

- **位置**：`cmd/note_v2.go:332-362`
- **现象**：dry-run 分支手动构造 `protocol.PlanChange` 列表，并返回顶层字段 `dry_run/applied/changed/plan_hash/plan`；非 dry-run 分支返回 `move` + `receipt`。
- **不合理原因**：
  - 其他 note 命令（create/append/patch/replace/delete）统一通过 `mutationResponse` 返回 `{"vault": ..., "result": <DryRunData|mutation>}`。
  - move 的 dry-run 与 apply 响应形状差异较大，调用方需要特殊解析，也容易出现 `changed` 字段语义不一致。
- **推荐方案**：将 move dry-run 也包装成 `{"vault": vault, "result": DryRunData, "plan_hash": ...}`，或让 `MovePlan` 实现一个返回 `DryRunData` 的方法，统一使用 `protocol.NewDryRunData`。
- **状态**：已确认。

### m5. `vault` 子命令 dry-run 未返回 vault 上下文

- **位置**：`cmd/vault_v2.go:179-183`、`220-232`、`258-267`、`318-322`
- **现象**：`vault add/remove/set-default/migrate` 的 dry-run 直接返回 `protocol.NewDryRunData(...)`，缺少 `vault` 字段；非 dry-run 分支均返回 `{"vault": vault, ...}`。
- **不合理原因**：与 `note`、`daily`、`metadata` 等命令 dry-run 返回 vault 上下文不一致；调用方在计划阶段无法确认目标 vault。
- **推荐方案**：dry-run 响应统一包含 `vault` 字段，例如 `map[string]any{"vault": vault, "result": dry}`。
- **状态**：已确认。

### m6. Obsidian 配置读取错误被静默忽略

- **位置**：`pkg/obsidian/config.go:29-40`、`46-60`、`65-76`
- **现象**：`ExcludedPaths`、`DefaultNoteFolder`、`ReadDailyNotesConfig` 在文件读取失败或 JSON 解析失败时返回 `nil` 或空结构体，不向上层报告任何 warning。
- **不合理原因**：`.obsidian/app.json` 或 `daily-notes.json` 损坏时，CLI 会回退到默认行为，用户/Agent 无法得知配置未生效。
- **推荐方案**：返回 warnings 列表（如 `ReadDailyNotesConfig` 改为返回 `(DailyNotesConfig, []string)`），并在 `daily.get/create/append` 的 JSON 输出中携带 `warnings`。
- **状态**：已确认。

### m7. `scripts/lint-skills.sh` 未校验 frontmatter `name` 与目录名一致性

- **位置**：`scripts/lint-skills.sh:66-73`
- **现象**：脚本校验 `name` 格式、`description`、不允许额外 key，但不比较 `name` 与实际目录名。
- **不合理原因**：`skills/evals/scenarios.json` 要求 skill `name` 与目录名对应；不一致会导致 manifest 与文件系统不匹配，调用方可能找不到 Skill。
- **推荐方案**：提取目录名 `dirname=$(basename "$(dirname "$file")")`，并断言 `name` 等于 `dirname`。
- **状态**：已确认。

### m8. `pkg/obsidian/discovery.go` Vault 发现路径未通过 `pathpolicy` 规范化

- **位置**：`pkg/obsidian/discovery.go:49`
- **现象**：`DiscoverObsidianVaultsFrom` 使用 `filepath.Clean(path)` 后直接返回。
- **不合理原因**：Vault 共同规范要求路径边界校验；发现结果后续会进入迁移/注册流程，若包含隐藏段或符号链接逃逸，可能在注册阶段才暴露。
- **推荐方案**：在发现阶段也使用 `pathpolicy.NewResolver(path)` 或 `pathpolicy` 的校验函数做规范化与边界检查。
- **状态**：已确认（影响较小，因注册阶段另有 `canonicalVaultPath`）。

## 5. 已排除 / 未确认的候选问题

| 候选问题 | 来源 | 排除原因 |
|---|---|---|
| `storage.ReadSnapshot` 跟随符号链接 | 子代理 | 实际先以 `os.Lstat` 检查 `Mode().IsRegular()`，符号链接会被拒绝，不会进入 `os.Open`。 |
| `note.append` 允许空 `--if-match` | 子代理 | `docs/spec/NOTE_OPERATIONS.md` 明确 `--if-match` 为可选；实现会在 plan 阶段读取当前 revision 并作为原子写入的 precondition。 |
| `skills/evals/scenarios.json` 使用 outcome `stale` | 子代理 | `scenarios.schema.json` 与 `cmd/skill_evals_test.go:233` 均显式允许 `stale`，非缺陷。 |
| `docs/compatibility.json` 状态为 `release-candidate` | 子代理 | P4-T06 明确说明在 P4-T05 移动端观察完成前保持 `release-candidate`，属于计划内状态。 |

## 6. 推荐处理优先级

1. **高**：修复 M1（配置原子写入）和 M2（List 隐藏文件过滤）。
2. **高**：清理 m2（V1 死代码），降低回归与误注册风险。
3. **中**：统一 m4、m5（dry-run 响应形状），提升调用方一致性。
4. **中**：补充文档与 lint：m3、m7。
5. **低**：改进配置错误提示 m6 和发现路径规范化 m8。
6. **低**：为 note create/move 的目录创建补齐路径策略 m1。

## 7. 声明

本报告基于 2026-07-27 的代码快照生成，**仅做记录，未改动任何源文件**。
