# obs

敏感操作的结构化审计留痕——字段值统一经 `guardx/logredact` 打码 —— oagentkit 模块矩阵 **tier 2**（集成层，依赖 tier0/tier1 + 重第三方地板）。

```
go get github.com/oagentkit/obs
```

| 包 | 职责 |
|---|---|
| `audit` | 敏感操作的结构化审计留痕（自 `guardx/audit` 迁入 v0.8.0）：字段值统一经 `logredact` 打码 |

> **本仓已降为 `audit` 单包仓。** 原 `langfuse` 包（Langfuse Public API 只读客户端）
> 已迁往 `github.com/oagentkit/llm/obsx` —— **读写同域**（写端本来就在 llm 仓的
> `obsx`：`otelgen.go` 的 OTel generation span 出口）⇒ **breaking**，见下方迁移须知。

## 下游迁移须知（breaking）

`obs/langfuse` 的**导出面整体迁走**（原包的 6 个导出符号全部在 llm 侧同名承接）：

```go
import "github.com/oagentkit/obs/langfuse"   // ❌ 已删除
```

改成

```go
import "github.com/oagentkit/llm/obsx"      // ✅ 读写同域
```

**符号名逐字未变**（`Client` / `NewClient` / `Query` / `Batch` / `Trace` /
`Observation`）⇒ **调用点零改动，只动 import 行**。

- 只用 `obs/langfuse` 的下游：迁走后本仓可以整条依赖摘掉（若仍用 `obs/audit` 则保留）。
- 本仓不再依赖 `oekit/pkg` 与 `oekit/concurrency` —— 那两条边只被 langfuse 客户端用到。
- 承接方是同批的 llm 侧 PR（两边一起落地，缺一不可）。

## 定位

**tier 2** — 集成层，依赖 tier0/tier1 + 重第三方地板。

- 本仓依赖的矩阵内仓：`guardx`（audit 用 `guardx/logredact` 打码）。

### 为什么 audit 留在本仓、langfuse 走

两者的**概念域不同**，不是「同属观测所以放一起」：

| | 概念域 | 归属 |
|---|---|---|
| `audit` | **执行面的操作留痕**：谁做了什么、参数是什么、结果如何 | 留在本仓 |
| `langfuse` 读端 | **观测数据出口**：把 trace 拉回来做观测报告 | 随写端去 `llm/obsx` |

- **写端在 llm 仓**：`obsx` 的 `otelgen.go` 是 Langfuse / Jaeger 侧 generation 视图的
  发射出口（v0.12.0 起自 `obs/obsx` 迁入 llm 仓）。读端与写端分居两仓时，
  「Langfuse 上有什么、是谁写进去的」没有一处能答全 ⇒ 读写同域。
- **同层依赖禁令**：矩阵里 `obs` 与 `llm` **同为 tier 2** ⇒ `obs` **不得**依赖 `llm`。
  所以**不能**在本仓留一个转发到 `llm/obsx` 的兼容壳（那会构成同层依赖，矩阵纪律禁止）
  ⇒ `langfuse` **直接删除**，调用方改 import 路径。

### 为什么 audit 的打码依赖 guardx

审计载荷常含平台 token、连接串、Bearer 头 ⇒ 字段值统一经
`guardx/logredact` 打码后再落。guardx 是 tier 0 零依赖底层，本仓 tier 2 ⇒
**依赖方向合法**（tier 2 → tier 0）。

## 外部依赖地板

- **零第三方依赖。** 只依赖 `github.com/oekit/logx` 与矩阵内的
  `github.com/oagentkit/guardx`。
  （eino 依赖已随 `obsx` 迁往 `llm` 仓而从本仓 go.mod 消失，见 CHANGELOG；
  `oekit/pkg` 与 `oekit/concurrency` 已随 `langfuse` 迁往 `llm/obsx` 而从本仓 go.mod 消失。）

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

本仓提供**执行面留痕**（audit），不提供 LLM 场景适配——
「把 LLM 调用挂到 OTel 上」那一层在 `github.com/oagentkit/llm/obsx`。
oekit 作为微服务工具链不做 LLM 场景；两者互补，不重叠。
