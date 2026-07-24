# P1-T01：配置边界与 Vault 发现

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`obs-cli`
- 依赖：P0-T01、P0-T02

## 目标

将 CLI 自有配置与 Obsidian 官方配置分离；官方配置仅用于发现和导入，不由 CLI 修改。

## 实施步骤

1. 盘点 `pkg/config` 当前读取和写入路径。
2. 定义 V2 自有配置位置、Schema 和版本字段。
3. 将读取 `obsidian.json` 封装为只读 discover provider。
4. 实现 `vault discover/list/get/add/remove/set-default`。
5. 明确 Vault ID、名称和规范路径的唯一性。
6. 配置更新使用文件锁或原子 compare-and-swap。
7. 提供旧配置一次性迁移或明确拒绝信息。

## 交付物

- V2 配置模型及迁移代码
- Vault 命令和 JSON 测试
- 配置与发现文档

## 验收标准

- [ ] CLI 不写入 Obsidian 的 `obsidian.json`。
- [ ] 多个 `open=true` Vault 的发现结果稳定。
- [ ] 重复名称和重复规范路径返回明确错误。
- [ ] 并发更新配置不会丢失数据。
- [ ] 配置解析失败不会静默回退。

## 验证

```bash
go test ./pkg/config ./pkg/obsidian -run 'Config|Discover|Vault'
go run . vault discover --output json
```

