# obs CHANGELOG

> 本文件由 `matrix-tools/scripts/gen_changelogs.py` 从 git 历史生成，按 tag 切段、取每段内的 conventional commit 渲染。

## [v0.7.0] — 2026-10-19

- 杂务(deps)：依赖地板对齐远端最新（textx v0.5.0） (13)

## [v0.6.0] — 2026-10-19

- 杂务(deps)：依赖地板对齐远端最新（textx v0.4.0 / concurrency v0.17.1 / logx v0.7.1 / pkg v0.15.1） (#13)

## [v0.5.0] — 2026-10-19

- 杂务(deps)：oekit 依赖升级（concurrency v0.16.0→v0.17.0）

## [v0.4.1] — 2026-10-18

- 与 v0.4.0 相比无代码差异（版本号/打包变更）

## [v0.4.0] — 2026-10-18

### 新增：obs 仓建立——LLM 可观测（2 包）

从 agentkit 单仓切出 tier1 能力层模块，module 路径 github.com/oagentkit/obs。
收编 2 个包：obsx（eino callbacks LLM 追踪 + token 用量 + 分阶段打标）/ langfuse（v3 只读客户端）
矩阵内依赖：textx。AGENTS.md 立了 trace 复用硬约束：必须复用 oekit/tracing 的 Provider
与 logx 的 trace 提取器，不自建 trace。
门禁：go build / go vet / go test -race 2 包全绿，gofmt 干净。

- 杂务(deps)：依赖地板升到「历史最新 + 0.1.0」 (#3)
- 杂务(deps)：跟进最新 tag (#2)
- 杂务(deps)：oekit 依赖升级 (#1)
- 杂务(deps)：oekit/concurrency v0.12.0 -> v0.13.1
- 杂务(deps)：oekit 三模块升至最新（pkg v0.13.0 / concurrency v0.12.0 / logx v0.7.0）
- 杂务(deps)：oekit/concurrency v0.6.0 -> v0.9.0

- CI：接入 org 共享门禁（规则本体在 oagentkit/.github）
