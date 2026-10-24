# obs

观测客户端 + 审计留痕——Langfuse 追踪客户端 / 敏感操作结构化审计 —— oagentkit 模块矩阵 **tier 2**（集成层，依赖 tier0/tier1 + 重第三方地板）。

```
go get github.com/oagentkit/obs
```

| 包 | 职责 |
|---|---|
| `langfuse` | Langfuse Public API 只读客户端（v3 self-hosted 契约面） |
| `audit` | 敏感操作的结构化审计留痕（自 `guardx/audit` 迁入 v0.8.0）：字段值统一经 `logredact` 打码 |

## 定位

**tier 2** — 集成层，依赖 tier0/tier1 + 重第三方地板。

- 本仓依赖的矩阵内仓：`guardx`。

## 外部依赖地板

- **零第三方依赖。** 只依赖 `github.com/oekit/logx`、`github.com/oekit/pkg`
  与矩阵内的 `github.com/oagentkit/guardx`。
  （eino 依赖已随 `obsx` 迁往 `llm` 仓而从本仓 go.mod 消失，见 CHANGELOG。）

## 与 oekit 的关系

本矩阵（oagentkit，Agent/LLM 编排工具链）**依赖** oekit（Open Easy Kit，Go 微服务工具链）：
`oekit/pkg` / `oekit/concurrency` / `oekit/logx` 等。oekit 是另一套独立矩阵，不是子品牌。

## 开发

```bash
go test ./... -race
go vet ./...
gofmt -l .
```

规范见 [AGENTS.md](AGENTS.md)。

## License

MIT
### trace 复用约束（不可自建）

**必须复用 `oekit/tracing` 的 Provider 和 `oekit/logx` 的 trace 提取器
（`logx.ExtractTrace`/`SetTraceExtractor`），不自建一套 trace。**
否则 agentkit 日志里的 trace_id 会与 OTel 的对不上，跨服务链路断裂。

### 与 oekit 的边界

本仓是**矩阵里唯一直接做 eino callbacks 适配**的仓 —— oekit 作为微服务工具链不会做 LLM 场景。
`oekit/tracing`（3 个导出）提供 Provider，本仓提供「把 LLM 调用挂上去」的适配层。两者互补，不重叠。