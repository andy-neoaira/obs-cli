# P1-T04：JSON 响应与稳定错误码

- 状态：`已完成`
- 负责人：`Codex`
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

- [x] JSON 模式 stdout 可直接被 JSON parser 读取。
- [x] 失败响应也满足 Schema。
- [x] `log.Fatal` 不存在于可复用业务包。
- [x] 相同领域错误跨命令返回相同 code。
- [x] request ID 在响应和诊断日志中保持一致。

## 验证

```bash
go test ./... -run 'JSON|Envelope|ErrorCode|ExitCode'
rg -n 'log\\.Fatal|os\\.Exit' pkg cmd
```

## 完成记录

- 完成日期：`2026-07-24`
- 新增 `pkg/protocol`，统一领域错误、成功/失败 envelope、warning、request ID、renderer 与退出码。
- `vault` V2 命令的成功、领域失败、参数失败、非法 request ID 和非法 output 均只向 stdout 输出一个 JSON object。
- 根执行器不再直接退出进程；已渲染的 V2 错误保持协议退出码，诊断日志携带相同 request ID。
- 所有 Cobra handler 均返回 error，`pkg` 与 `cmd` 中已清除 `log.Fatal` 和 `os.Exit`。
- 黄金成功/失败输出从 `response-v2.schema.json` 动态读取约束进行合同验证。
- Schema 增加与退出码 10 对应的 `INTERNAL_ERROR`，避免未分类错误产生不符合 Schema 的失败响应。
- 全量测试、全量 Race、Vet、license check 与协议专项测试通过。
