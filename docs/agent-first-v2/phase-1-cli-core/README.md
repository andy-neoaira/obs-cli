# P1：obs-cli Agent-first 安全内核

## 阶段目标

将现有 Human-first CLI 破坏性升级为可被 Agent 稳定调用的 V2 操作内核。

## 进入条件

- P0-T01 至 P0-T05 已完成。
- CLI 协议、Vault 规范和并发协议已经冻结。

## 任务进度

阶段进度：`4 / 8`

- [x] [P1-T01 配置边界与 Vault 发现](./P1-T01-config-and-discovery.md)
- [x] [P1-T02 安全路径解析](./P1-T02-safe-paths.md)
- [x] [P1-T03 原子写入与 Revision 内核](./P1-T03-atomic-storage.md)
- [x] [P1-T04 JSON 响应与稳定错误码](./P1-T04-json-errors.md)
- [ ] [P1-T05 Capabilities、通用参数与 Dry-run](./P1-T05-capabilities-dry-run.md)
- [ ] [P1-T06 Note 原子操作 API](./P1-T06-note-operations.md)
- [ ] [P1-T07 Move 与链接重写事务](./P1-T07-move-link-transaction.md)
- [ ] [P1-T08 V2 命令树、质量门禁与发布迁移](./P1-T08-v2-release.md)

推荐顺序：P1-T01/P1-T02/P1-T04 → P1-T03 → P1-T05 → P1-T06 → P1-T07 → P1-T08。

## 阶段完成标准

- [ ] 所有 Agent 核心命令支持 JSON。
- [ ] 所有文件写入使用统一原子存储层。
- [ ] 所有更新型操作支持 revision 前置条件。
- [ ] 路径和符号链接不能逃逸 Vault。
- [ ] 批量或多文件修改不会静默部分成功。
- [ ] Human-first 隐式行为已删除或移出核心命令。
- [ ] 单元、竞态、覆盖率和发布检查通过。

## 阶段验证

```bash
cd /Users/andy/github/obs-cli
GOCACHE=/private/tmp/obs-cli-gocache go test ./...
GOCACHE=/private/tmp/obs-cli-gocache go test -race ./...
go vet ./...
gofmt -l .
```
