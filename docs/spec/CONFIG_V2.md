# obs-cli V2 配置

- Schema 版本：`2`
- 默认文件：操作系统用户配置目录下的 `obs-cli/config-v2.json`
- 所有者：`obs-cli`
- Obsidian 官方配置：只读发现来源

## 配置边界

`obs-cli` 不再将 Vault 注册写入 Obsidian 的 `obsidian.json`。两类文件的职责如下：

| 文件 | 读 | 写 | 用途 |
|---|---:|---:|---|
| `obs-cli/config-v2.json` | 是 | 是 | CLI Vault registry、默认 Vault、CLI 偏好 |
| Obsidian `obsidian.json` | 是 | 否 | `vault discover` 与一次性迁移来源 |
| Vault `.obsidian/*.json` | 是 | 否 | Vault 行为发现，例如 Daily Notes |

## Schema

```json
{
  "version": 2,
  "default_vault_id": "vlt_0123456789abcdef0123456789abcdef",
  "default_open_type": "editor",
  "vaults": {
    "vlt_0123456789abcdef0123456789abcdef": {
      "id": "vlt_0123456789abcdef0123456789abcdef",
      "name": "Personal",
      "path": "/absolute/canonical/path/Personal"
    }
  }
}
```

约束：

- `version` 必须为 `2`，未知版本明确拒绝。
- Vault map key 必须与记录中的 `id` 一致。
- 名称按 Unicode case-insensitive 比较后必须唯一。
- 规范路径按保守的 case-insensitive 比较后必须唯一。
- 配置中的路径必须是绝对且已清理的路径；通过 CLI 注册或迁移时，还会验证目录存在并解析符号链接后再写入。
- `default_vault_id` 必须指向已注册记录。
- `default_open_type` 只能为空、`obsidian` 或 `editor`。
- 配置存在但 JSON 或约束无效时返回错误，不回退为空配置。

## 写入安全

配置更新使用 `config-v2.json.lock` 对协作的 CLI 进程串行化。持锁后重新读取最新配置，执行修改并验证完整 Schema，再把同目录临时文件 flush 后原子 rename。

锁超时会失败，不会绕过锁覆盖。异常遗留锁需要用户确认没有运行中的 CLI 后手工处理；实现不按固定时间静默删除锁。

## 命令

```text
obs-cli vault discover
obs-cli vault list
obs-cli vault get <id-or-name>
obs-cli vault add <path> [--name <name>] [--set-default]
obs-cli vault remove <id-or-name>
obs-cli vault set-default <id-or-name>
obs-cli vault migrate
```

所有新 `vault` 命令输出 `obs-cli/v2` JSON envelope。`discover` 只读官方配置，并稳定地把 open Vault 排在 closed Vault 前。

## 旧配置迁移

`vault migrate` 是一次性显式操作：

1. 严格解析旧 `preferences.json`；损坏时停止，不静默忽略。
2. 只读发现 Obsidian 注册的 Vault。
3. 只导入当前存在的目录，并生成 CLI 自有稳定 ID。
4. 旧默认名称只有唯一匹配时才转换成 `default_vault_id`。
5. 重复名称、重复规范路径或已含 Vault 的 V2 配置会拒绝迁移。
6. 写入 `migrated_from`；无法唯一恢复默认值时写入 `migration_warning`。

迁移不会修改或删除旧配置与 Obsidian 配置。
