# obs-cli V1 配置

- Schema 版本：`1`
- 默认文件：操作系统用户配置目录下的 `obs-cli/config.json`
- 所有者：`obs-cli`
- Obsidian 官方配置：只读发现来源

测试、沙箱或多实例运行可通过绝对路径环境变量
`OBS_CLI_CONFIG_HOME` 覆盖“用户配置目录”这一层。最终文件仍位于
`$OBS_CLI_CONFIG_HOME/obs-cli/config.json`。相对路径会被拒绝，生产环境
未设置该变量时行为不变。

## 配置边界

`obs-cli` 不再将 Vault 注册写入 Obsidian 的 `obsidian.json`。两类文件的职责如下：

| 文件 | 读 | 写 | 用途 |
|---|---:|---:|---|
| `obs-cli/config.json` | 是 | 是 | CLI Vault registry 与默认 Vault |
| Obsidian `obsidian.json` | 是 | 否 | `vault discover` 的只读来源 |
| Vault `.obsidian/*.json` | 是 | 否 | Vault 行为发现，例如 Daily Notes |

## Schema

```json
{
  "version": 1,
  "default_vault_id": "vlt_0123456789abcdef0123456789abcdef",
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

- `version` 必须为 `1`，未知版本明确拒绝。
- Vault map key 必须与记录中的 `id` 一致。
- 名称按 Unicode case-insensitive 比较后必须唯一。
- 规范路径按保守的 case-insensitive 比较后必须唯一。
- 配置中的路径必须是绝对且已清理的路径；通过 CLI 注册时，还会验证目录存在并解析符号链接后再写入。
- `default_vault_id` 必须指向已注册记录。
- 配置存在但 JSON 或约束无效时返回错误，不回退为空配置。

## 写入安全

配置更新使用 `config.json.lock` 对协作的 CLI 进程串行化。持锁后重新读取最新配置，执行修改并验证完整 Schema，再把同目录临时文件 flush 后原子 rename。

锁超时会失败，不会绕过锁覆盖。异常遗留锁需要用户确认没有运行中的 CLI 后手工处理；实现不按固定时间静默删除锁。

## 命令

```text
obs-cli vault discover
obs-cli vault list
obs-cli vault get <id-or-name>
obs-cli vault add <path> [--name <name>] [--set-default]
obs-cli vault remove <id-or-name>
obs-cli vault set-default <id-or-name>
```

所有新 `vault` 命令输出 `obs-cli/v1` JSON envelope。`discover` 只读官方配置，并稳定地把 open Vault 排在 closed Vault 前。

运行时只读取 `config.json`，不探测、复制或迁移其他历史配置文件。新注册表通过
`vault discover`、`vault add` 和 `vault set-default` 显式建立。
