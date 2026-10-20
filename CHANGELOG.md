# obs CHANGELOG

> 本文件由 `matrix-tools/scripts/gen_changelogs.py` 从 git 历史生成，按 tag 切段、取每段内的 conventional commit 渲染。

## [v0.9.2] — 2026-10-20

### 修复(go)：Go 地板降到 go 1.24.0，CI go-version 同步到 1.24

承接 matrix-tools#22（tier0 三仓的地板被无谓抬高）——本轮把 tier1+ 里
**依赖链全兼容**的仓也降下来（matrix-tools#34 的排查结论）。
## 依据
- **依赖链上没有要求更高版本的包**：本仓的每个依赖（直接与间接）的 go 指令
  都 ≤ 1.24.x，本仓声明 1.26.0 没有任何依赖需要它。
- **代码本身兼容 1.24**：真实 go1.24.0 工具链（`GOTOOLCHAIN=go1.24.0`）
  build + test 全绿——实跑，不是推断。
- **门槛是 go 指令不是 toolchain**（编译器实证）：`go 1.28.0` 被拒，
  `go 1.24.0` + `toolchain go1.28.0` 通过。
## toolchain 行的取舍：删掉
toolchain **不构成门槛**，但 `toolchain go1.27.2` 会让 `go` 命令**自动下载**
1.27.2，与「最低支持版本是 1.24」相悖。删掉，也与基座 oekit 六个基础仓一致。
## 依赖地板同步重钉
降 go 指令后，`require` 必须指向**已发新 tag 的 tier0 仓**（它们的 tag 快照
里才是 go 1.24.0）——否则 MVS 会拉到旧 tag 的 1.26 地板，降级形同虚设。
本仓的内部 require 已随 `go mod tidy` 一并抬到对应的新 tag。
## 改动
- `go.mod`：`go 1.26.0` → `go 1.24.0`；删 `toolchain go1.27.2`；
  内部 require 抬到 tier0 新 tag
- `.github/workflows/ci.yml`：`go-version: '1.26'` → `'1.24'`
CI 的 go-version 与 go.mod 的 go 指令**成对**——这是跨层通用纪律，不因本次
降级而放松。CI 用高版本构建、go.mod 承诺低版本时，CI 全绿但用户按最低版本
编译失败，门禁绿了而承诺没兑现。
## 验证
- `GOTOOLCHAIN=go1.24.0 go build ./...` → exit 0
- `GOTOOLCHAIN=go1.24.0 go test ./...` → 全绿
- `go mod tidy` → 零差异
门禁见 oagentkit/matrix-tools：`scripts/verify_go_floor.py`（判据 4 覆盖
tier1+ 相对地板，判据独立不复用任何生成器）。

## [v0.9.1] — 2026-10-20

### 修复(docs)：外部依赖地板改为零第三方依赖——eino 随 obsx 迁出已不在 go.mod

Co-authored-by: Claude Opus 4.8 (1M context) <noreply@anthropic.com>

### 修复(go.sum)：清掉移除 eino 后残留的 oekit/core 条目

obs 迁出 obsx 后不再依赖 eino，但 go.sum 里 oekit/core（eino 传递进来的）
只剩一条 /go.mod 哈希行留着。CI 的 `go mod tidy -diff` 判红。
依赖内容没变（v0.9.0 的包与 API 不变），故不发新版本——已发布物不动，
下游 go.sum 校验和不受影响。

## [v0.9.0] — 2026-10-20

### 重构：obsx 迁出本仓 —— eino 耦合的观测适配层该待在重仓

BREAKING CHANGE: github.com/oagentkit/obs/obsx → github.com/oagentkit/llm/obsx
obsx 是 eino callbacks 适配层（intrinsically 绑 eino 框架）。它在本仓时，
**langfuse 与 audit 这两个轻包被白拖 27 个第三方模块**（eino + sonic 全家）——
Go 的依赖面按 module 结算。
迁到 llm（本仓早已 import obsx 做成本记账与阶段打标，同一关注点）后：
- obs 的 eino 依赖从 go.mod 消失 → langfuse / audit 只剩 logx 级别的地板
- llm 仓的「模型调用观测」与「模型客户端」住在一起
本仓剩两包：langfuse（Langfuse 追踪客户端）、audit（审计留痕）。
关联 oagentkit/guardx#3

## [v0.8.0] — 2026-10-19

### 新增(audit)：接收 guardx/audit —— 审计留痕落到本来就带 logx 的仓

自 github.com/oagentkit/guardx/audit 迁入（guardx v0.7.0 起该路径退役）。
为什么迁到 obs：audit 依赖 oekit/logx（审计日志门面），而本仓本来就带 logx。
guardx 是 tier 0 零依赖底层 —— 依赖面按 module 结算，那条 require 会顺着
依赖树扩散到矩阵每一层。带重依赖的能力该待在带该依赖的仓里。
包名与职责不变（审计留痕：结构化审计事件，字段值统一经 logredact 打码），
只是 import 路径从 github.com/oagentkit/guardx/audit 变为
github.com/oagentkit/obs/audit。
关联 oagentkit/guardx#3

## [v0.7.0] — 2026-10-19

- 杂务(deps)：依赖地板对齐远端最新（textx v0.5.0） (13)

## [v0.6.0] — 2026-10-19

- 杂务(deps)：依赖地板对齐远端最新（textx v0.4.0 / concurrency v0.17.1 / logx v0.7.1 / pkg v0.15.1）

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

- 杂务(deps)：依赖地板升到「历史最新 + 0.1.0」
- 杂务(deps)：跟进最新 tag
- 杂务(deps)：oekit 依赖升级
- 杂务(deps)：oekit/concurrency v0.12.0 -> v0.13.1
- 杂务(deps)：oekit 三模块升至最新（pkg v0.13.0 / concurrency v0.12.0 / logx v0.7.0）
- 杂务(deps)：oekit/concurrency v0.6.0 -> v0.9.0

- CI：接入 org 共享门禁（规则本体在 oagentkit/.github）
