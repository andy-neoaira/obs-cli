---
name: obsidian-vault-setup
version: 1.0.0
description: "初始化和配置 obs-cli Vault：注册 vault、设置默认 vault、设置默认打开方式、查看当前配置。"
metadata:
  category: "obsidian"
  requires:
    bins: ["obs-cli"]
---

# obsidian-vault-setup

用于通过 `obs-cli` 初始化或调整 Obsidian Vault 配置。

## 适用场景

当用户表达以下意图时使用：

- 帮我配置 obs-cli
- 把这个目录注册为 Obsidian vault
- 设置默认 vault
- 设置默认用 editor / Obsidian 打开
- 查看当前已注册 vault 或默认 vault

## 需要收集的信息

- `vault_path`：要注册的 vault 目录，可选。
- `set_default`：是否设为默认 vault。
- `open_type`：默认打开方式，只能是 `obsidian` 或 `editor`。
- `vault`：要设为默认的 vault 名称或路径，可选。

## 操作流程

### 1. 先查看现有配置

```bash
obs-cli list-vaults
obs-cli list-vaults --default
```

如果用户只想查看配置，不要继续执行写入操作。

### 2. 注册新 Vault

用户提供明确路径时：

```bash
obs-cli add-vault "<vault_path>"
```

如果用户要求同时设为默认：

```bash
obs-cli add-vault "<vault_path>" --set-default
```

### 3. 设置默认 Vault

```bash
obs-cli set-default-vault "<vault-name>"
```

### 4. 设置默认打开方式

```bash
obs-cli set-default-vault --open-type editor
obs-cli set-default-vault --open-type obsidian
```

### 5. 输出最终状态

```bash
obs-cli list-vaults --default
```

## 安全规则

- 注册或修改默认配置前，确保用户已经明确给出目标路径、vault 名称或打开方式。
- `open_type` 只能使用 `obsidian` 或 `editor`，不要猜其他值。
- `remove-vault` 不属于本 skill 的默认流程；除非用户明确要求移除 vault，否则不要调用。
- 如果 `obs-cli` 不存在，提示用户先构建或安装本项目，例如 `make build` 或 `go install github.com/andy-neoaira/obs-cli@latest`。

## 输出格式

完成后汇报：

- 当前默认 vault 名称
- 当前默认 vault 路径
- 当前默认打开方式
- 后续常用命令示例，例如 `obs-cli daily`、`obs-cli search-content "关键词"`
