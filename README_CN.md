# obs-cli

**中文** | [🇺🇸 English](./README.md)

> **注意**：obs-cli **无需 Obsidian 正在运行**即可工作，非常适合脚本化、自动化和纯终端环境。

---

## 目录

- [项目描述](#项目描述)
- [Agent-first V2 架构](#agent-first-v2-架构)
- [安装](#安装)
  - [下载预编译二进制](#下载预编译二进制)
  - [从源码编译](#从源码编译)
  - [无图形界面 / 未安装 Obsidian](#无图形界面--未安装-obsidian)
- [使用指南](#使用指南)
  - [帮助信息](#帮助信息)
  - [编辑器标志](#编辑器标志)
  - [添加 Vault](#添加-vault)
  - [移除 Vault](#移除-vault)
  - [列出 Vault](#列出-vault)
  - [设置默认 Vault 和打开方式](#设置默认-vault-和打开方式)
  - [打开笔记](#打开笔记)
  - [每日笔记](#每日笔记)
  - [搜索笔记](#搜索笔记)
  - [搜索笔记内容](#搜索笔记内容)
  - [列出 Vault 内容](#列出-vault-内容)
  - [打印笔记](#打印笔记)
  - [创建 / 更新笔记](#创建--更新笔记)
  - [移动 / 重命名笔记](#移动--重命名笔记)
  - [删除笔记](#删除笔记)
  - [Frontmatter](#frontmatter)
- [排除文件](#排除文件)
- [参与贡献](#参与贡献)
- [许可证](#许可证)

---

## 项目描述

Obsidian 是一款功能强大且可扩展的知识库应用，基于本地纯文本笔记文件夹工作。这款 CLI 工具（使用 Go 语言编写）让你可以通过终端与 Obsidian 交互。目前支持打开、搜索、列出、移动、创建、更新和删除笔记等操作。

---

## Agent-first V2 架构

本项目正在向 Agent-first V2 演进：`obs-cli` 将成为供 AI Agent 和场景化 Skills 使用的安全、机器可读执行层。

- Markdown Vault 是唯一内容事实源。
- Obsidian、`obs-cli` 与 [`miniobsidian.nvim`](https://github.com/andy-neoaira/miniobsidian.nvim) 是操作同一 Vault 的同级客户端。
- `miniobsidian.nvim` 不强制依赖 `obs-cli`；未来的可选集成只用于通过 CLI 协议协商的高级能力。
- `obs-cli` 不依赖 Neovim；它只读发现或显式导入 Obsidian 所有的配置，不取得这些配置的所有权。
- V2 核心采用非交互设计，不提供 TTY 问答、内置 picker、自动启动编辑器或隐式打开应用。

完整决策见 [ADR-001：Agent-first 产品边界与三入口架构](./docs/architecture/ADR-001-agent-first-boundary.md)，配置边界见 [V2 配置规范](./docs/spec/CONFIG_V2.md)，路径边界见 [Vault 安全路径策略](./docs/spec/PATH_POLICY.md)，执行任务见 [V2 联合改造计划](./docs/agent-first-v2/README.md)。

共同规范状态：`target_contract = vault-contract/v1`，`implemented_contract = null`。目标规则见 [Vault 共同约定](./docs/spec/VAULT_CONVENTIONS.md)；只有 P1 共享 fixture 通过后才会声明实现符合该版本。

> 在 V2 迁移任务完成前，本文后续命令仍描述当前 CLI。

---

## 安装

### 下载预编译二进制

最简单的安装方式是从 [GitHub Releases](https://github.com/andy-neoaira/obs-cli/releases) 页面下载预编译好的二进制文件。

**支持的平台：**

| 操作系统 | 架构 | 发布包文件名 |
|---|---|---|
| macOS (Universal) | amd64 + arm64 | `obs-cli_darwin_all.tar.gz` |
| Linux | amd64 | `obs-cli_linux_amd64.tar.gz` |
| Linux | arm64 | `obs-cli_linux_arm64.tar.gz` |
| Windows | amd64 | `obs-cli_windows_amd64.tar.gz` |
| Windows | arm64 | `obs-cli_windows_arm64.tar.gz` |

**一键安装（macOS / Linux）：**

```bash
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  darwin) asset_os="darwin"; asset_arch="all" ;;
  linux) asset_os="linux"; asset_arch="$arch" ;;
  *) echo "Unsupported OS: $os" >&2; exit 1 ;;
esac

case "$asset_arch" in
  x86_64) asset_arch="amd64" ;;
  aarch64) asset_arch="arm64" ;;
esac

curl -sL -o obs-cli.tar.gz "https://github.com/andy-neoaira/obs-cli/releases/latest/download/obs-cli_${asset_os}_${asset_arch}.tar.gz"
tar xzf obs-cli.tar.gz

# 移动到 PATH 中的目录
mkdir -p ~/bin
mv obs-cli ~/bin/
```

> **注意：** Apple Silicon Mac 上 `uname -m` 输出 `arm64`，Intel Mac 上输出 `x86_64`。

**Windows（PowerShell）：**

```powershell
# 下载最新版本
Invoke-WebRequest -Uri "https://github.com/andy-neoaira/obs-cli/releases/latest/download/obs-cli_windows_amd64.tar.gz" -OutFile "obs-cli.tar.gz"

# 解压
tar xzf obs-cli.tar.gz

# 添加到 PATH（可选）
$env:PATH += ";$PWD"
```

### 从源码编译

需要 [Go](https://go.dev/dl/) 1.19 或更高版本。

#### 快速编译

```bash
git clone https://github.com/andy-neoaira/obs-cli.git
cd obs-cli
go build -o obs-cli .
sudo install -m 755 obs-cli /usr/local/bin/
```

#### 使用 Make

本项目包含 `Makefile`，提供了便捷的开发构建和发布目标：

```bash
# 为所有支持的平台构建二进制文件（Darwin、Linux、Windows）
make build-all

# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 安装 git 钩子
make install-hooks
```

#### 交叉编译

Go 内置支持交叉编译。设置 `GOOS` 和 `GOARCH` 环境变量即可为不同平台构建：

```bash
# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o obs-cli-darwin-amd64 .

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o obs-cli-darwin-arm64 .

# Linux
GOOS=linux GOARCH=amd64 go build -o obs-cli-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -o obs-cli-linux-arm64 .

# Windows
GOOS=windows GOARCH=amd64 go build -o obs-cli-windows-amd64.exe .
```

#### 开发环境搭建

```bash
# 1. 克隆仓库
git clone https://github.com/andy-neoaira/obs-cli.git
cd obs-cli

# 2. 下载依赖（使用 vendor 模式）
go mod vendor

# 3. 运行测试确保一切正常
go test ./...

# 4. 编译二进制文件
go build -o obs-cli .

# 5. 运行 CLI
./obs-cli --help
```

#### 从源码安装

编译完成后，你可以将二进制文件安装到系统中：

```bash
# Linux / macOS
sudo cp obs-cli /usr/local/bin/

# 或者使用 go install（安装到 $GOPATH/bin 或 $HOME/go/bin）
go install github.com/andy-neoaira/obs-cli@latest

# 验证安装
obs-cli --version
```

### 无图形界面 / 未安装 Obsidian

如果你在无图形界面的服务器上运行，或者没有安装 Obsidian（例如服务器环境、容器或没有 GUI 的系统），你仍然可以使用此 CLI。当 Obsidian 已安装时，它会自动注册 vault。对于无图形界面环境，你需要通过 CLI 手动注册。

**设置步骤：**

```bash
# 注册你的 vault 目录
obs-cli add-vault /home/user/vaults/my-brain

# 将其设为默认
obs-cli set-default-vault "my-brain"

# 或者一步完成
obs-cli add-vault /home/user/vaults/my-brain --set-default
```

多 vault 设置：

```bash
obs-cli add-vault /home/user/vaults/personal
obs-cli add-vault /home/user/vaults/work
obs-cli set-default-vault "personal"
```

之后你可以通过 `--vault "work"` 来指定特定的 vault。

<details>
<summary>手动设置（不使用 CLI 命令）</summary>

1. 创建 Obsidian 配置目录：
   ```bash
   mkdir -p ~/.config/obsidian
   ```

2. 创建 `obsidian.json` 文件并写入你的 vault 配置：
   ```json
   {
     "vaults": {
       "any-unique-id": {
         "path": "/home/user/vaults/my-brain"
       }
     }
   }
   ```
   键（`any-unique-id`）可以是任意值。CLI 使用**目录名称**作为 vault 名称（例如上面的 `my-brain`）。请使用**绝对路径**，不要使用 `~`，因为 CLI 不会将其展开为主目录。

</details>

---

## 使用指南

### 帮助信息

```bash
# 查看所有命令说明
obs-cli --help
```

### 编辑器标志

`open`、`daily`、`search`、`search-content`、`create` 和 `move` 命令支持 `--editor`（或 `-e`）标志，可以在你的默认文本编辑器中打开笔记，而不是在 Obsidian 应用中打开。这对于快速编辑或在纯终端环境中工作时非常有用。

编辑器由 `EDITOR` 环境变量决定（例如 `"vim"`、`"code"` 或 `"code -w"`）。如果未设置，默认使用 `vim`。

**支持的编辑器：**

- 终端编辑器：vim、nano、emacs 等
- 带等待标志的 GUI 编辑器：VSCode (`code`)、Sublime Text (`subl`)、Atom、TextMate
  - CLI 会自动为支持的 GUI 编辑器添加 `--wait` 标志，确保它们在你关闭文件之前保持阻塞状态

**示例：**

```bash
# 设置你的首选编辑器（添加到 ~/.zshrc 或 ~/.bashrc 以永久生效）
export EDITOR="code"  # 或 "vim"、"nano"、"subl" 等

# 在支持编辑器标志的命令中使用
obs-cli open "note.md" --editor
obs-cli daily --editor
obs-cli search --editor
obs-cli search-content "term" --editor
obs-cli create "note.md" --open --editor
obs-cli move "old.md" "new.md" --open --editor
```

若想避免每次传递 `--editor`，可以一次性配置默认打开方式：

```bash
obs-cli set-default-vault --open-type editor
```

### 添加 Vault

将目录注册为 Obsidian vault。如果 Obsidian 配置文件（`~/.config/obsidian/obsidian.json`）不存在，会自动创建。

如果你已安装 Obsidian，打开 vault 时会自动注册。只有在无图形界面环境或 Obsidian 未安装的情况下（服务器、容器、CI），才需要使用此命令。

```bash
# 注册一个 vault
obs-cli add-vault /path/to/vault

# 注册并设为默认
obs-cli add-vault /path/to/vault --set-default
```

### 移除 Vault

按 vault 名称从 Obsidian 配置中移除一个 vault。不会删除磁盘上的任何文件。如果被移除的 vault 是默认 vault，默认设置会被清除。

```bash
# 按 vault 名称移除
obs-cli remove-vault "{vault-name}"

```

### 列出 Vault

列出所有已注册的 Obsidian vault。默认 vault 会标注 `(default)`。

```bash
# 列出所有 vault（名称和路径，默认标注）
obs-cli list-vaults

# 以 JSON 格式输出
obs-cli list-vaults --json

# 只输出路径（适合脚本使用）
obs-cli list-vaults --path-only

# 只显示默认 vault（名称、路径、打开方式）
obs-cli list-vaults --default

# 只获取默认 vault 路径（适合脚本使用）
obs-cli list-vaults --default --path-only
```

你可以将以下内容添加到 shell 配置文件（如 `~/.zshrc`），以便快速导航到默认 vault：

```bash
obs_cd() {
    local result=$(obs-cli list-vaults --default --path-only)
    [ -n "$result" ] && cd -- "$result"
}
```

然后你可以在终端中使用 `obs_cd` 快速跳转到默认 vault 目录。

### 设置默认 Vault 和打开方式

设置默认 vault 和/或默认打开方式，供后续命令使用。如果没有设置默认 vault，可以在其他命令中通过 `--vault` 标志指定。

```bash
# 按名称设置默认 vault
obs-cli set-default-vault "{vault-name}"

# 设置默认打开方式：'obsidian'（默认）或 'editor'
obs-cli set-default-vault --open-type editor

# 同时设置两者
obs-cli set-default-vault "{vault-name}" --open-type editor
```

当 `default_open_type` 设为 `editor` 时，支持 `--open` 的命令会自动在 `$EDITOR` 中打开笔记，无需每次传递 `--editor`。

注意：`open` 和其他命令使用 vault 的基目录作为工作目录，而不是你终端的当前工作目录。

### 打开笔记

在 Obsidian（或你的默认编辑器）中打开指定笔记。笔记参数必须是 vault 内相对路径，`.md` 后缀可省略。

```bash
# 在默认 vault 中打开笔记
obs-cli open "{note-name}"

# 在指定 vault 中打开笔记
obs-cli open "{note-name}" --vault "{vault-name}"

# 打开笔记并定位到特定标题（区分大小写）
obs-cli open "{note-name}" --section "{heading-text}"

obs-cli open "{note-name}" --vault "{vault-name}" --section "{heading-text}"

# 在默认编辑器中打开笔记（而非 Obsidian）
obs-cli open "{note-name}" --editor
```

### 每日笔记

直接在磁盘上创建或打开今天的每日笔记。**Obsidian 不需要正在运行**。如果 vault 中存在 `.obsidian/daily-notes.json`，CLI 会从中读取 `folder`、`format`（Moment.js 日期格式，默认 `YYYY-MM-DD`）和 `template`。创建新的每日笔记时会使用模板文件的内容。如果配置缺失或无法读取，则使用默认值（vault 根目录、`YYYY-MM-DD`、无模板）。

```bash
# 在默认 vault 中创建/打开每日笔记
obs-cli daily

# 在指定 vault 中创建/打开每日笔记
obs-cli daily --vault "{vault-name}"

# 在默认编辑器中创建/打开每日笔记
obs-cli daily --editor

# 向每日笔记添加内容（如果笔记已存在则追加）
obs-cli daily --content "abcde"

# 从文件安全追加多行内容
obs-cli daily --content-file ./daily-entry.md

# 从 stdin 安全追加内容
printf '%s\n' "## Work Log" "- shipped feature" | obs-cli daily --content-file -

# 添加内容并在编辑器中打开
obs-cli daily --content "abcde" --editor
```

### 搜索笔记

启动模糊搜索，在终端中显示 vault 中的笔记。按回车即可在 Obsidian 中打开选中的笔记。

```bash
# 在默认 vault 中搜索
obs-cli search

# 在指定 vault 中搜索
obs-cli search --vault "{vault-name}"

# 搜索并在默认编辑器中打开选中的笔记
obs-cli search --editor
```

### 搜索笔记内容

搜索笔记内容中包含指定关键词的笔记。默认会打开交互式选择器，让你在 Obsidian（或编辑器）中打开选中的笔记。对于自动化和脚本，使用 `--no-interactive` 或 `--format json` 将结果打印到标准输出。

```bash
# 在默认 vault 中搜索内容
obs-cli search-content "search term"

# 在指定 vault 中搜索内容
obs-cli search-content "search term" --vault "{vault-name}"

# 搜索并在默认编辑器中打开选中的笔记
obs-cli search-content "search term" --editor

# 以 grep 风格打印结果到标准输出（非交互式）
obs-cli search-content "search term" --no-interactive

# 以 JSON 格式打印结果（适合脚本，隐含非交互模式）
obs-cli search-content "search term" --format json

# 分页结果（默认每页 25 条，最大 100 条）
obs-cli search-content "search term" --format json --page 1 --page-size 50
```

### 列出 Vault 内容

列出 vault 路径中的文件和文件夹。如果不提供路径，则列出 vault 根目录。

```bash
# 列出 vault 根目录
obs-cli list

# 列出默认 vault 中某个子文件夹的内容
obs-cli list "001 Notes"

# 列出指定 vault 中某个子文件夹的内容
obs-cli list "001 Notes" --vault "{vault-name}"
```

### 打印笔记

打印指定笔记的内容。笔记参数必须是 vault 内相对路径，`.md` 后缀可省略。

```bash
# 打印默认 vault 中的笔记
obs-cli print "{note-name}"

# 按路径打印默认 vault 中的笔记
obs-cli print "{note-path}"

# 打印指定 vault 中的笔记
obs-cli print "{note-name}" --vault "{vault-name}"
```

### 创建 / 更新笔记

直接在磁盘上创建笔记。**Obsidian 不需要正在运行**。如果笔记已存在且未传递 `--overwrite` 或 `--append`，命令会失败。中间目录会自动创建。

当笔记名称没有显式路径（不含 `/`）时，CLI 会读取 vault 中的 `.obsidian/app.json`，检查是否配置了默认文件夹（`newFileLocation: "folder"` 和 `newFileFolderPath`）。如果已配置，笔记会放在该文件夹中。如果配置缺失或无法读取，笔记将创建在 vault 根目录。

```bash
# 在默认 vault 中创建空笔记
obs-cli create "{note-name}"

# 在指定 vault 中创建空笔记
obs-cli create "{note-name}" --vault "{vault-name}"

# 创建带内容的笔记
obs-cli create "{note-name}" --content "abcde"

# 从文件创建多行内容笔记
obs-cli create "{note-name}" --content-file ./note.md

# 从 stdin 创建笔记
printf '%s\n' "# Title" "Body text" | obs-cli create "{note-name}" --content-file -

# 覆盖已有笔记
obs-cli create "{note-name}" --content "abcde" --overwrite

# 向已有笔记追加内容
obs-cli create "{note-name}" --content "abcde" --append

# 创建笔记并在 Obsidian 中打开
obs-cli create "{note-name}" --content "abcde" --open

# 创建笔记并在默认编辑器中打开
obs-cli create "{note-name}" --content "abcde" --open --editor
```

### 移动 / 重命名笔记

移动笔记（vault 顶层的路径）到新名称（vault 顶层的路径）。如果路径相同但名称不同，则视为重命名。vault 中所有指向该笔记的链接都会自动更新。

```bash
# 在默认 vault 中重命名笔记
obs-cli move "{current-note-path}" "{new-note-path}"

# 在指定 vault 中重命名笔记
obs-cli move "{current-note-path}" "{new-note-path}" --vault "{vault-name}"

# 重命名并在 Obsidian 中打开
obs-cli move "{current-note-path}" "{new-note-path}" --open

# 重命名并在默认编辑器中打开
obs-cli move "{current-note-path}" "{new-note-path}" --open --editor
```

### 删除笔记

删除指定笔记（vault 顶层的路径）。

```bash
# 删除默认 vault 中的笔记
obs-cli delete "{note-path}"

# 删除指定 vault 中的笔记
obs-cli delete "{note-path}" --vault "{vault-name}"
```

### Frontmatter

查看和修改笔记中的 YAML frontmatter。笔记参数必须是 vault 内相对路径，`.md` 后缀可省略。

```bash
# 打印笔记的 frontmatter
obs-cli frontmatter "{note-name}" --print

# 编辑 frontmatter 字段（不存在则创建）
obs-cli frontmatter "{note-name}" --edit --key "status" --value "done"

# 删除 frontmatter 字段
obs-cli frontmatter "{note-name}" --delete --key "draft"

# 在指定 vault 中使用
obs-cli frontmatter "{note-name}" --print --vault "{vault-name}"
```

---

## 排除文件

CLI 尊重 Obsidian 的**排除文件**设置（`设置 > 文件与链接 > 排除的文件`）。

- `search` — 被排除的笔记不会出现在模糊搜索器中
- `search-content` — 被排除的文件夹不会被搜索
- `print` / `frontmatter` — 使用准确的 vault 内相对路径时仍可访问被排除的文件

其他显式路径命令（`open`、`move`、`delete` 等）在你提供准确路径时仍然可以访问被排除的文件。

---

## 发布版本

本项目使用 [GoReleaser](https://goreleaser.com/) 自动构建并发布二进制文件到 [GitHub Releases](https://github.com/andy-neoaira/obs-cli/releases)。

**通过推送版本标签触发发布：**

```bash
# 1. 提交你的更改
git add .
git commit -m "feat: 你的更改说明"

# 2. 打语义化版本标签
git tag v0.1.0

# 3. 推送标签（这会触发发布工作流）
git push origin master
git push origin v0.1.0
```

GitHub Actions 工作流将自动执行：
1. 运行测试
2. 为所有支持的平台构建二进制文件（Darwin、Linux、Windows，amd64 + arm64）
3. 创建 GitHub Release 并提供可下载的压缩包

只有匹配 `v*.*.*` 格式的标签推送才会触发发布。普通分支推送不会触发。

---

## 参与贡献

Fork 本项目，添加你的功能或修复，然后提交 Pull Request。你也可以在 [Issues](https://github.com/andy-neoaira/obs-cli/issues/new/choose) 中报告 bug 或请求新功能。

---

## 多语言文档

- [English](./README.md)
- [简体中文](./README_CN.md)

---

## 许可证

本项目派生自 [Yakitrak/notesmd-cli](https://github.com/Yakitrak/notesmd-cli)。[LICENSE](./LICENSE) 保留了上游 MIT 版权与许可声明，并追加了本项目修改部分的版权声明。

项目采用 [MIT 许可证](./LICENSE)。Vendored 依赖的许可证与声明汇总在 [THIRD_PARTY_NOTICES.md](./THIRD_PARTY_NOTICES.md)，并随发布归档一起分发。
