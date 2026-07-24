# P1-T04：JSON 响应与稳定错误码

- 状态：`未开始`
- 负责人：`待分配`
- 涉及项目：`obs-cli`
- 依赖：P0-T03

## 目标

让 CLI 的所有 Agent 核心操作严格遵守 `obs-cli/v2` 输出和错误协议。

## 实施步骤

1. 建立领域错误类型，禁止通过字符串匹配判断错误。
2. 建立统一 response renderer。
3. stdout 只写最终数据；日志、警告和调试信息写 stderr。
4. Cobra 命令返回 error，由根命令统一映射退出码。
5. 为每个稳定错误码建立单元测试。
6. 使用 JSON Schema 验证命令黄金输出。
7. 提供 `--output json`；Agent 核心命令默认 JSON 的决策按协议执行。

## 交付物

- 领域错误包
- JSON renderer
- 协议一致性测试

## 验收标准

- [ ] JSON 模式 stdout 可直接被 JSON parser 读取。
- [ ] 失败响应也满足 Schema。
- [ ] `log.Fatal` 不存在于可复用业务包。
- [ ] 相同领域错误跨命令返回相同 code。
- [ ] request ID 在响应和诊断日志中保持一致。

## 验证

```bash
go test ./... -run 'JSON|Envelope|ErrorCode|ExitCode'
rg -n 'log\\.Fatal|os\\.Exit' pkg cmd
```

