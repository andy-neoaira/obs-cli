# V1-T04：重置配置文件与 Schema 版本

- 状态：`实现完成，待提交`
- 优先级：`高`
- 涉及项目：`obs-cli`
- 依赖：V1-T01、V1-T02、V1-T03

## 目标

建立唯一的首版配置：`obs-cli/config.json`、根类型 `Config`、内容
`"version": 1`。正式运行时不读取、不探测、不迁移 `config-v2.json`。

## 目标格式

```json
{
  "version": 1,
  "default_vault_id": "vlt_...",
  "vaults": {
    "vlt_...": {
      "id": "vlt_...",
      "name": "Personal",
      "path": "/absolute/path"
    }
  }
}
```

## 修改范围

- `pkg/config/constants.go`
- `pkg/config/cli_path.go`
- `pkg/config/store.go`
- 配置和 Registry 测试
- Vault 命令测试
- 文档中的配置路径和 Schema
- RC smoke 使用的隔离配置

## 硬切换规则

- 唯一文件名为 `config.json`。
- 唯一 Schema 版本为 `1`。
- 不 dual-read。
- 不自动 rename/copy `config-v2.json`。
- 不在缺少 `config.json` 时探测旧文件。
- 不通过 warning 静默回退到旧配置。
- 开发者本机旧配置由仓库外一次性操作处理，不进入正式代码。

## legacy import 决策

删除 `vault migrate`、旧 `preferences.json` 导入实现、对应 operation、测试和迁移文档。
新注册表只通过 `vault discover`、`vault add` 和显式配置操作建立，不提供 legacy import。

## 执行步骤

1. 修改失败测试，使其期望 `config.json` 和 `version: 1`。
2. 将配置文件常量改为 `config.json`。
3. 将 `CurrentConfigVersion` 改为 `1`。
4. 将唯一路径函数定为 `ConfigPath`。
5. 更新默认 Store 和 Registry 构造路径。
6. 更新所有测试 fixture 和临时配置文件名。
7. 增加负向测试：
   - 只有 `config-v2.json` 时按“新配置不存在”处理；
   - `version: 2` 明确返回 unsupported version；
   - 不产生自动迁移副作用；
   - dry-run 不创建配置。
8. 验证并发更新、原子替换、权限和目录 fsync 行为不回归。
9. 更新当前配置规范。
10. 删除所有旧配置常量和路径函数。

## 验收标准

- [ ] 唯一默认配置路径为 `obs-cli/config.json`。
- [ ] 新配置固定写入 `"version": 1`。
- [ ] `version != 1` 明确拒绝。
- [ ] 只有旧文件时不会自动读取或转换。
- [ ] 不存在旧配置常量、路径函数或 alias。
- [ ] 原子写入、锁、权限、并发更新测试通过。
- [ ] dry-run 保持零配置副作用。

## 验证命令

```bash
go test ./pkg/config ./pkg/obsidian ./cmd -run 'Config|Vault|Registry|DryRun'
go test -race ./pkg/config
rg -n 'config-v2\.json|ObsCLIV2ConfigFile|V2Path|CurrentConfigVersion *= *2' \
  cmd pkg scripts testdata docs/spec
git diff --check
```

检索必须无输出。

## 本地数据处理

执行者若需要保留个人测试注册表，应在仓库外手工转换，并先备份。该操作不是项目安装或
运行流程的一部分，也不得写入正式脚本。

## 回滚

整体回滚配置提交。回滚后删除任务期间生成的临时测试配置；不得自动把真实用户
`config.json` 覆盖回旧文件名。

## 完成记录

- 完成日期：`2026-07-28（工作区实现）`
- 提交：`未提交`
- 配置负向测试：`待填写`
- legacy import 删除证据：`待填写`
