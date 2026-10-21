# AGENTS.md — oagentkit/obs

> 观测客户端 + 审计留痕——Langfuse 追踪客户端 / 敏感操作结构化审计
> module：`github.com/oagentkit/obs` · org：[oagentkit = Open Agent Kit](https://github.com/oagentkit)
> 矩阵分层：**tier 1**（能力层，依赖 tier0）

## 本仓包含

```
`langfuse/`
`audit/`      ← 自 guardx 迁入（v0.8.0）：审计留痕，依赖 oekit/logx
```

## 硬规则

### 分层与依赖

- 依赖严格自底向上，**禁止反向依赖**。本仓是 tier 1，只允许依赖 tier ≤ 1 的仓。
- 本仓依赖的矩阵内仓：`textx`、`guardx`（audit 用 `guardx/logredact` 打码）。
  ⚠️ 两者都是 tier 0，本仓 tier 1 —— 依赖方向合法。
- 改 module 路径必须同步：全部 import、go.mod、README 安装行。

### 错误处理

- 一律使用 `github.com/oekit/pkg/errors`；业务错误用 BizError。
- 所有 error 禁止 `_` 忽略。

### 并发安全

- 禁止裸 `go func`：用 `github.com/oekit/concurrency/async` 的 GoSafe 系列（panic 被 recover）。
- 共享状态用 mutex / atomic 保护；包级全局可变状态用 `atomic.Pointer[T]`。

### API 设计

- 函数选项模式命名 `XxxOption`（禁止裸 `Option`，多包 co-import 会冲突）。
- 导出符号 PascalCase + godoc 注释；包文档放 `doc.go`。

### 日志

- 日志走 `oekit/logx`，组件 logger 用 `logx.Delegating("oagentkit-obs")`。
- **观测数据不得携带凭据原文**，落盘前过 `guardx/logredact`。
- `audit` 包的 logger 为 nil 时回退 `logx.NewAuditLogger`（级别锁 Info、不受
  `LOG_LEVEL` 降级、SetLevel 到 WARN/ERROR 被拒）—— 这是刻意设计，别改回普通 logger。

### 发版顺序

本仓的 `CHANGELOG.md` 由 `matrix-tools/scripts/gen_changelogs.py` 从 git 历史生成，
**不要手改**。发版顺序的**单一副本**见 org 规则仓
[CONTRIBUTING §发版](https://github.com/oagentkit/.github/blob/main/CONTRIBUTING.md#发版)：
**预写待发版段 → 提交 CHANGELOG.md → 在那个提交上打 annotated tag**。
tag 必须落在 CHANGELOG 提交**之后** —— 下游 `go get` 拿到的是 tag 快照，
快照里的 CHANGELOG.md 若缺自己那一段，发布说明就是残缺的。

⚠️ 无参重生成 CHANGELOG 会**吞掉待发版段**：生成器只按已存在的 tag 切段，
`--next-version` 是唯一能产出待发版段的开关。**预写之后、打 tag 之前，
不要跑无参重生成。**

打 tag 前自检依赖地板：内部 `require` 要钉在**打 tag 那一刻**各依赖仓的最新 tag
（`matrix-tools/scripts/verify_tag_floor.py`）。本仓已有 3 个 CHANGELOG-at-tag 欠账
（`obs@v0.8.0`/`v0.9.0`/`v0.9.1`）—— **反面例子，不是范例**。

## 仓特有纪律

### trace 复用约束（不可自建）

**必须复用 `oekit/tracing` 的 Provider 和 `oekit/logx` 的 trace 提取器
（`logx.ExtractTrace`/`SetTraceExtractor`），不自建一套 trace。**
否则 agentkit 日志里的 trace_id 会与 OTel 的对不上，跨服务链路断裂。

### 与 oekit 的边界

本仓是**矩阵里唯一直接做 eino callbacks 适配**的仓 —— oekit 作为微服务工具链不会做 LLM 场景。
`oekit/tracing`（3 个导出）提供 Provider，本仓提供「把 LLM 调用挂上去」的适配层。两者互补，不重叠。

## 明确不属于本仓的

| 能力 | 归属 |
|---|---|
| 文本/差异/会话/报告 | `github.com/oagentkit/textx` |
| 事件总线/流控/子进程/热插拔/钩子 | `github.com/oagentkit/runtime` |
| 脱敏/出口围栏/权限/沙箱解析/审计 | `github.com/oagentkit/guardx` |
| DB 队列与选主 | `github.com/oagentkit/queue` |
| 通用 errors/stringx/encoding/mathx | `github.com/oekit/pkg` |
| goroutine 池/单飞/重试/令牌桶 | `github.com/oekit/concurrency` |
