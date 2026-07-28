# Vault Conventions

- 规范标识：`vault-contract/v1`
- 状态：已实现并通过双仓共享 fixture
- 发布日期：2026-07-24
- 适用实现：`obs-cli` V1、`miniobsidian.nvim`
- 架构依据：[ADR-001](../architecture/ADR-001-agent-first-boundary.md)

## 0. 规范语言与兼容性

### VC-0.1 规范关键词

本文中的“必须”“禁止”“应该”“可以”分别对应 MUST、MUST NOT、SHOULD、MAY。带 `VC-*` 的条目是稳定规则 ID，测试、错误和变更记录应该引用规则 ID。

### VC-0.2 版本格式

规范版本使用 `vault-contract/v<major>[.<minor>]`：

- 破坏现有合法 Vault 解释、Note ID、链接目标或写入结果的变更，必须升级主版本。
- 只增加可选能力且旧实现可以安全忽略的变更，可以升级次版本。
- 文字澄清不得改变可观察行为。

### VC-0.3 符合性声明

实现必须分别声明：

- `target_contract`：计划实现的规范版本。
- `implemented_contract`：已经通过共享 fixture 的规范版本；尚未通过时为 `null`。
- 已知偏差列表。

文档不得把“目标版本”表述为“已经完整支持”。

## 1. Vault 与路径

### VC-1.1 Vault 根目录

Vault 根目录是一个存在的目录。实现必须将输入根目录解析为绝对规范路径，并解析根目录自身的符号链接。内容操作不得依赖当前工作目录隐式选择 Vault。

`.obsidian/` 的存在可以作为发现信号，但显式注册的无 `.obsidian/` Markdown 目录是否可用由客户端能力决定；同一客户端必须稳定返回该能力。

### VC-1.2 逻辑路径

协议和共同 fixture 中的 Vault 相对逻辑路径：

- 必须使用 `/` 作为分隔符。
- 必须相对 Vault 根目录。
- 禁止为空（表示根目录的专用操作除外）。
- 禁止以 `/`、盘符、UNC 前缀或 `~` 开头。
- 禁止 `.`、`..`、空路径段和 NUL。
- 禁止以 `/` 结尾表示文件。

客户端可以接受平台原生分隔符作为人类输入，但进入协议响应前必须转为逻辑路径。

### VC-1.3 路径规范化

实现必须先做语法规范化，再做文件系统边界检查。不得只用字符串前缀判断路径是否位于 Vault 中。

对于已经存在的路径，必须解析所有符号链接并验证最终真实路径仍在真实 Vault 根目录内。对于新路径，必须解析最近的已存在父目录，再验证剩余路径段。

### VC-1.4 符号链接

符号链接目标逃逸 Vault 时，所有读取、写入、扫描和附件操作都必须拒绝，错误应引用 `VC-1.4`。

指向 Vault 内部的符号链接可以用于读取；如果多个逻辑路径指向同一物理文件，发现结果必须标记为物理身份冲突，修改操作禁止静默选择其中一个路径。

### VC-1.5 平台大小写与 Unicode

- POSIX 通常大小写敏感，实现必须保留实际文件名大小写。
- macOS 文件系统可能大小写敏感或不敏感，禁止根据操作系统名称假设，必须以实际文件系统结果为准。
- Windows 默认大小写不敏感，协议逻辑路径仍保留实际显示大小写。
- 实现不得在写回时擅自改变 Unicode normalization form。
- 同一目录存在仅大小写或 Unicode normalization 不同、但在目标平台无法稳定区分的名称时，必须报告冲突或歧义。

共享 fixture 使用 UTF-8 路径；Windows fixture 不创建保留设备名、尾随空格或尾随点路径。

### VC-1.6 Windows 特殊路径

Windows 实现必须拒绝盘符绝对路径、UNC 路径、alternate data stream 语法和系统保留设备名。跨平台创建操作应该提前拒绝无法在 Windows 正常检出的名称，并返回可解释错误。

### VC-1.7 新目录

创建笔记或附件时可以创建缺失的父目录，但必须在通过 VC-1.2 至 VC-1.6 后执行。目录创建失败不得留下目标文件。

## 2. 内容范围与忽略规则

### VC-2.1 Markdown 文件

新笔记必须使用小写 `.md` 扩展名。扫描时可以识别大小写不同的 Markdown 扩展名，但同一 Note ID 对应多个扩展名大小写变体时必须报告冲突。

目录、符号链接目录和非 Markdown 文件不是笔记。

### VC-2.2 内部目录

默认笔记扫描、搜索、补全、审计和批量修改必须排除：

- `.obsidian/`
- `.git/`、`.hg/`、`.svn/`
- `.trash/`、`.Trash/`
- `.obs-cli/`

内部目录中的文件不得成为普通 Note ID。工具可以通过专用配置读取接口只读访问 `.obsidian/`。

### VC-2.3 隐藏路径

除 VC-2.2 的内部目录外，以 `.` 开头的路径段默认不进入发现、搜索和批量操作。显式读取是否允许由 capability 声明；显式写入默认拒绝，除非调用方使用专用、可审计的允许选项。

### VC-2.4 Obsidian Ignore Filters

`.obsidian/app.json` 的 `userIgnoreFilters` 应用于发现、搜索、反向链接、审计和批量变更。被忽略目标的显式读取可以允许，但响应必须标记 `ignored: true`；显式修改必须要求调用方确认允许 ignored target。

无效 ignore 配置必须产生 warning，不得导致客户端扫描 Vault 外部内容。

### VC-2.5 模板与附件范围

模板 `.md` 文件仍是可寻址内容，但默认搜索结果应标记 `kind: template`。附件不是笔记，不参与 Note ID 和正文全文搜索。

## 3. 编码、换行与 Revision 输入

### VC-3.1 文本编码

新建 Markdown 必须使用 UTF-8 且不写 BOM。读取到 UTF-8 BOM 时可以解析，但未修改 BOM 的操作必须保留它。无法作为 UTF-8 解码的文件必须报告编码错误，不得自动转码覆盖。

### VC-3.2 换行

新文件默认使用 LF。局部更新已有文件时必须保留其主换行风格（LF 或 CRLF）和“文件末尾是否有换行”的状态，除非操作明确要求格式化。

混合换行文件可以被读取，但写入前必须返回 warning；不得在无关局部更新中全文件归一化。

### VC-3.3 原始字节

revision 基于磁盘原始字节计算，不基于解析后的文本、换行归一化结果或 Frontmatter map。具体算法由并发协议定义。

## 4. Note ID、标题与 Frontmatter

### VC-4.1 Note ID

Note ID 是去掉最后 `.md` 扩展名后的 Vault 相对逻辑路径，保留目录和文件名大小写。

示例：

```text
Projects/Agent CLI.md → Projects/Agent CLI
```

标题、Frontmatter `id` 字段和文件内容都不改变 Note ID。

### VC-4.2 Note 引用输入

机器协议必须优先接受完整 Note ID 或带 `.md` 的完整逻辑路径。裸 basename 只有在允许的搜索范围内唯一匹配时才能解析；零匹配返回 not found，多匹配返回 ambiguous，禁止静默使用第一个结果。

### VC-4.3 显示标题

需要显示标题时按以下顺序选择：

1. 有效 Frontmatter 中非空字符串 `title`。
2. Markdown 正文第一个一级标题文本。
3. Note ID 的最后一个路径段。

显示标题只用于 UI，不用于机器身份或无条件目标解析。

### VC-4.4 Frontmatter 边界

YAML Frontmatter 只有在可选 BOM 之后的第一个逻辑行是 `---`，并存在独立结束分隔行时才成立。正文中的 `---` 不得被解析成 Frontmatter。

### VC-4.5 Frontmatter 修改

- Frontmatter 是可选的，不要求所有笔记拥有固定 Schema。
- 字段级修改必须保留未知字段的语义和值。
- Frontmatter 无效时，字段级修改必须失败并返回解析位置；不得用空 map 覆盖。
- 删除最后一个字段后，可以删除整个空 Frontmatter 块，但行为必须在 dry-run 中可见。
- 键名大小写敏感，不得自动合并名称相似的未知字段。

### VC-4.6 自动字段

客户端不得仅因打开、读取或移动笔记而自动增加 `title`、`date`、`tags` 或其他 Frontmatter。创建时使用的默认字段属于模板或显式创建策略，必须在 plan 中可见。

## 5. Wikilink 与 Markdown Link

### VC-5.1 Wikilink 结构

Wikilink 的基本结构为：

```text
[[target]]
[[target|alias]]
[[target#heading]]
[[target#heading|alias]]
[[target#^block-id|alias]]
```

解析器必须分别保留 target、alias、heading 和 block ID。alias 和 fragment 不属于 Note ID。

### VC-5.2 Wikilink 目标解析

解析顺序：

1. target 含目录时，按 Vault 根目录相对 Note ID 精确解析。
2. target 不含目录时，在允许范围内按 basename 搜索。
3. 精确大小写优先；大小写折叠匹配只有唯一时才允许。
4. 多个候选始终返回 `AMBIGUOUS_NOTE` 及候选列表。
5. 不存在时由调用场景决定提示创建或返回 not found，不得自动创建错误目录中的同名笔记。

### VC-5.3 Fragment

解析出目标文件后再解析 heading 或 block。heading 采用 Markdown 可见文本匹配；存在重复 heading 时必须按文档顺序和 Obsidian heading anchor 规则消歧。block ID 必须精确匹配。

找不到 fragment 不等于找不到笔记，响应必须区分两者。

### VC-5.4 Alias

alias 只影响显示文本。移动、重命名或补全不得把 alias 当成目标，也不得在无必要时改写 alias。

### VC-5.5 Markdown Link

Markdown 文件链接相对于包含链接的笔记目录解析；以 `/` 开头的 Vault 绝对风格链接是否启用必须由 capability 声明。URL、`mailto:`、data URI 和其他外部 scheme 不参与 Vault 文件重写。

路径中的 percent encoding 必须在解析时安全解码一次，禁止通过双重解码形成路径逃逸。

### VC-5.6 链接扫描边界

Wikilink 和 Markdown Link 只有在 Markdown 正文的链接语法位置才参与反向链接和重写。 fenced code、inline code、HTML comment 和 Frontmatter 标量中的相似文本不得被无上下文字符串替换。

### VC-5.7 新链接生成

当 basename 在目标范围内唯一时，可以生成短 Wikilink；存在或可能产生同名歧义时，必须生成包含目录的 Note ID。生成器必须保留调用方选择的 alias 和 fragment。

## 6. 附件

### VC-6.1 附件目录

附件目录必须是通过 VC-1 路径验证的 Vault 相对目录。空值表示由调用场景明确选择，不得隐式写到进程当前目录。

### VC-6.2 附件命名

附件文件名必须去除路径分隔符和 NUL，并遵守目标平台限制。目标已存在时默认失败；可以通过明确的唯一化策略生成新名称，但禁止静默覆盖。

### VC-6.3 附件写入

附件写入必须先落到目标目录中的临时文件，再原子替换为最终名称。读取剪贴板或源文件失败时不得创建空附件。

### VC-6.4 附件链接

标准 Markdown 附件链接默认相对于当前笔记目录生成，并使用 `/`。Wikilink 附件遵守 VC-5 的唯一解析规则。生成链接前必须验证从笔记位置解析后指向刚写入的附件。

### VC-6.5 附件移动

移动笔记默认不移动附件。移动附件或重写附件链接必须是独立、可预览的计划，且不得改写外部 URL 或同名的其他附件。

## 7. Daily Note 与模板

### VC-7.1 配置来源

存在有效 `.obsidian/daily-notes.json` 时，folder、format 和 template 是默认权威配置，CLI 与插件必须产生相同的有效配置。

配置文件不存在时可以使用 VC-7.2 默认值；文件存在但无法读取、JSON 损坏或字段值
不受支持时，不得静默退回默认路径执行写入。Agent CLI 必须返回包含配置文件相对路径
和失败类别的结构化错误。

客户端私有 override 只有在用户显式配置时才允许，并必须在结果或 health 中标记与 Obsidian 配置的偏差。

### VC-7.2 默认配置

官方配置不存在或字段为空时使用：

- folder：Vault 根目录
- format：`YYYY-MM-DD`
- template：无

客户端不得分别使用不一致的隐式默认目录。

### VC-7.3 日期与时区

Daily Note 的日期在操作开始时计算一次。使用调用环境的本地时区，响应或调试信息应包含日期、UTC offset 和所用格式。测试必须固定时间和时区。

“昨天”和“明天”按日历日期计算，不得用固定 `±86400` 秒计算。

### VC-7.4 日期格式

共同输入格式使用 Obsidian/Moment 风格。客户端必须维护受支持 token 表；遇到不支持 token 时返回明确错误或 capability warning，不得生成含错误替换的文件名。

格式结果必须通过 VC-1 路径验证，因为日期格式可以包含目录分隔符。

### VC-7.5 模板定位

template 按 Vault 相对 Note ID 解析，缺少 `.md` 时补充一次。模板路径必须通过 VC-1，多个候选或路径逃逸必须失败。配置了模板但模板不存在或不可读时，创建操作默认失败，不得静默创建另一种内容。

### VC-7.6 共同模板变量

`vault-contract/v1` 的共同变量为：

- `{{date}}`
- `{{time}}`
- `{{title}}`
- `{{filename}}`
- `{{yesterday}}`
- `{{tomorrow}}`
- `{{date:FORMAT}}`

变量大小写不敏感。未知变量保留原文并返回 warning。变量替换规则和支持 token 必须在两个实现的共享 fixture 中一致。

### VC-7.7 创建与已存在行为

Daily Note 目标不存在时，先渲染模板，再按原子写入协议创建。没有模板时创建空文件；客户端私有默认 Frontmatter 只有显式启用时才可加入。

目标已存在时，单纯“打开/获取今日笔记”不得修改内容。追加或更新必须是显式操作，并携带 revision 前置条件。

### VC-7.8 一致性要求

在相同 Vault 配置、时间、时区和输入下，`obs-cli` 与 `miniobsidian.nvim` 必须得到相同的 Daily Note 逻辑路径和初始文件字节。差异必须由 fixture 测试阻止。

## 8. 创建、移动、删除与链接重写

### VC-8.1 创建

创建目标已存在时默认失败。只有显式 replace 操作并满足 revision 前置条件时才允许替换。创建计划必须展示目标路径、模板、初始 Frontmatter 和附件副作用。

### VC-8.2 移动

移动前必须验证源存在、目标合法、目标不存在、父目录可创建，并解析源/目标物理身份。仅大小写改名在大小写不敏感文件系统上必须使用安全的中间步骤。

### VC-8.3 链接重写计划

移动后的链接重写必须基于解析后的链接 token 和实际目标身份生成计划。计划至少包含：

- 源和目标 Note ID。
- 每个受影响文件的 revision。
- 每处链接的旧 token、新 token 和位置。
- 不修改的歧义或无法解析链接。

### VC-8.4 多文件应用

应用移动计划前必须重新检查所有 revision。任一前置条件失败时默认不开始写入。执行中失败必须回滚；无法完整回滚时返回机器可读 partial failure 和恢复清单。

### VC-8.5 删除

删除必须是显式操作，并要求目标 revision。默认策略应使用可恢复的隔离区或备份；永久删除必须是单独 capability 和显式授权。

删除前应报告指向目标的有效链接，但是否允许存在 backlink 时删除由调用策略决定。

### VC-8.6 不相关内容

创建、移动、删除和链接重写不得格式化或重排不相关 Frontmatter、正文、换行和空白。

## 9. 外部修改与缓存

### VC-9.1 外部修改

Obsidian、同步服务、CLI 和 Neovim 都可能在客户端缓存之后修改文件。客户端不得把缓存内容视为最新磁盘状态；更新前必须重新读取或校验 revision。

### VC-9.2 Neovim Buffer

磁盘被外部修改时：

- 未修改 buffer 可以按用户配置重载，但必须触发正常的 Neovim 文件变更流程。
- 已修改 buffer 禁止自动重载，必须提示冲突并保留内存与磁盘两个版本。
- 需要合并时应该提供 base、buffer、disk 三方比较。

### VC-9.3 缓存失效

外部创建、删除、移动和内容修改必须在有界时间内使笔记、补全、反向链接和预览缓存失效。缓存可以提高性能，但不得绕过写入前 revision 校验。

### VC-9.4 同步中间态

读取到临时文件、零字节中间态或同步冲突副本时不得自动覆盖。客户端应短暂重试稳定读取或报告 transient error；冲突副本作为独立文件处理，不得擅自合并。

## 10. 配置读取与可观察性

### VC-10.1 Obsidian 配置只读

两个项目可以读取 `.obsidian/app.json`、`.obsidian/daily-notes.json` 等公开文件配置，但不得以普通设置操作修改它们。修改 Obsidian 官方配置必须属于未来独立、显式授权的 capability。

### VC-10.2 无效配置

配置文件存在但 JSON 无效时，不得伪装成“未配置”。操作必须返回错误或带来源路径的 warning，并根据操作安全性决定停止。

路径、Daily Template 和 Ignore Filters 等影响读取/写入范围的配置无效时，修改操作必须停止。

### VC-10.3 结果可观察性

涉及文件的机器响应至少应返回逻辑路径；读取和修改操作还应返回 revision。发生歧义时返回候选逻辑路径，不暴露 Vault 外部绝对路径。

## 11. 最小符合性矩阵

| 能力 | 必须通过的规则 |
|---|---|
| Vault 发现与路径 | VC-1、VC-2、VC-10 |
| Note 读写 | VC-1、VC-3、VC-4、VC-8、VC-9 |
| Wikilink/Backlink | VC-4、VC-5、VC-8 |
| 附件 | VC-1、VC-3、VC-6、VC-9 |
| Daily Note | VC-1、VC-3、VC-4、VC-7、VC-9 |
| Neovim buffer 协同 | VC-3、VC-9 |

实现只有在对应共享 fixture 全部通过后，才能声明该能力符合 `vault-contract/v1`。
