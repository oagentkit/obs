# obs CHANGELOG

> 本文件由 `matrix-tools/scripts/gen_changelogs.py` 从 git 历史生成，按 tag 切段、取每段内的 conventional commit 渲染。

## [v0.9.3] — 2026-10-20

- 杂务(deps)：依赖地板抬到最新

地板对齐到发版时刻的各依赖最新 tag：
- oekit/pkg、oekit/concurrency 抬到 v0.18.0（oekit/logx 已是最新 v0.7.1）
- 矩阵内跨仓依赖抬到各仓最新 tag
依据 matrix-tools#44：最新 tag 的依赖地板在「打 tag 那一刻」就已落后一格，
只能靠后续版本补齐（tag 受 protect-tags ruleset 保护不能重打）。


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

matrix-tools 的 verify_dep_floors.py 首次全矩阵扫描查出的地板漂移，本仓是受影响仓之一。
本轮按 tier 拓扑顺序推进：tier0（textx / runtime / guardx / queue）先发，
本仓再把地板跟到它们的新 tag——否则就是「发了新的、旧的立刻变落后」，永远收敛不了。
所升版本与原钉版本相比源码零改动，逐版实测确认：
- textx v0.4.0→v0.5.0 / runtime v0.6.0→v0.7.0 / guardx v0.4.1→v0.5.0
  全树逐字节相同（diff -rq 退出码 0）
- oekit/pkg v0.14.0→v0.15.1  仅新增 urlguard 包（纯标准库，pkg 的 go.mod
  连 require 块都没有，零新增三方依赖）
- oekit/concurrency v0.17.0→v0.17.1  仅三个文档文件
- oekit/logx v0.7.0→v0.7.1  仅 CI 配置与文档
无 API 变更、无兼容风险。
本地门禁：go mod tidy -diff 无输出 / go build / go vet / gofmt / go test -race 全通过。
关联 oagentkit/matrix-tools#13


## [v0.6.0] — 2026-10-19

- 杂务(deps)：依赖地板对齐远端最新（textx v0.4.0 / concurrency v0.17.1 / logx v0.7.1 / pkg v0.15.1）

matrix-tools 的 verify_dep_floors.py 首次全矩阵扫描查出：43 条第一方依赖里
28 条落后，横跨 11 个仓；本仓是其中之一。
所升版本与本仓原钉版本相比源码零改动，逐版实测确认：
- textx v0.3.0→v0.4.0 / obs v0.4.1→v0.5.0 / runtime v0.5.0→v0.6.0 /
  mcpkit v0.5.1→v0.6.0  全树逐字节相同（diff -rq 退出码 0）
- oekit/pkg v0.14.0→v0.15.1  仅新增 urlguard 包（纯标准库，pkg 的 go.mod
  连 require 块都没有，零新增三方依赖）
- oekit/concurrency v0.17.0→v0.17.1  仅三个文档文件
- oekit/logx v0.7.0→v0.7.1  仅 CI 配置与文档
无 API 变更、无兼容风险。
本地门禁：go mod tidy -diff 无输出 / go build / go vet / gofmt / go test -race 全通过。


## [v0.5.0] — 2026-10-19

- 杂务(deps)：oekit 依赖升级（concurrency v0.16.0→v0.17.0）

本 PR 含两项改动。
## 1. oekit 依赖地板
姊妹矩阵 oekit 的 `concurrency` 出了 v0.17.0，依赖地板一并对齐。
升级前已核对 `oekit/concurrency` v0.16.0...v0.17.0 的全部 2 个提交：
- `feat(ratelimit): KeyedLimiter.Create 护栏内原子登记 (Fixes #33)` — `ratelimit/keyed.go` +28 行
- `docs: 矩阵总览表 DESC 口径同步` — 仅 AGENTS.md / README.md / CHANGELOG.md
无 breaking / remove 标记，未触及任何已有导出签名，故为纯补丁式升级。
## 2. 迁出被 force 移动过的 tag
本仓原先依赖的姊妹仓 tag（发布后被强推到了重写后的 main）内容已与
sum.golang.org 记录不一致，于是任何走 `proxy.golang.org` 或 `GOPROXY=direct`
的构建都会 SECURITY ERROR。本地往往不报错——模块缓存预热 + goproxy.cn
还留着移动前的字节，把问题掩盖了。
已为 8 个模块重打不可变的**附注 tag**（内容与被移动 tag 指向同一 commit，
只换版本号，让校验和数据库重新记录一次正确哈希），并把依赖迁过去。
`go.sum` 的 h1 全部取自 sum.golang.org，仅就地替换目标模块的两行
（zip h1 + go.mod h1），其余行不动。


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

本轮做了版本线重置：清空全部历史 tag，按「历史最新 + 0.1.0」重打新版本
（如 llm v0.6.0 → v0.7.0、toolgate v0.6.1 → v0.7.0、agentkit v1.3.0 → v1.4.0）。
各仓 go.mod 的矩阵内依赖地板同步跟进到新版本号。
**go.sum 的 oagentkit 行已整段删除**，由 CI 的 `go mod tidy` 重算：
新 tag 的内容与任何历史版本都不同，`h1` 必须现算；本地没有 Go 工具链，
且 `/go.mod` 的 dirhash 无法自行复现（zip hash 可复现，gomod hash 不可复现）。
重算依赖 `.github` 的 go-ci 配了 `GOPRIVATE=github.com/oagentkit/*` 直连 VCS：
goproxy.cn 缓存着旧版本，若走 proxy 会拉回**旧代码**并算出旧 hash，
门禁全绿但依赖是错的。
（第二个提交会补上重算好的 go.sum。）
* chore(deps): 依赖地板升到「历史最新 + 0.1.0」
chore(deps): 依赖地板升到「历史最新 + 0.1.0」
本轮做了版本线重置：清空全部历史 tag，按「历史最新 + 0.1.0」重打新版本
（如 llm v0.6.0 → v0.7.0、toolgate v0.6.1 → v0.7.0、agentkit v1.3.0 → v1.4.0）。
各仓 go.mod 的矩阵内依赖地板同步跟进到新版本号。
**go.sum 的 oagentkit 行已整段删除**，由 CI 的 `go mod tidy` 重算：
新 tag 的内容与任何历史版本都不同，`h1` 必须现算；本地没有 Go 工具链，
且 `/go.mod` 的 dirhash 无法自行复现（zip hash 可复现，gomod hash 不可复现）。
重算依赖 `.github` 的 go-ci 配了 `GOPRIVATE=github.com/oagentkit/*` 直连 VCS：
goproxy.cn 缓存着旧版本，若走 proxy 会拉回**旧代码**并算出旧 hash，
门禁全绿但依赖是错的。
（第二个提交会补上重算好的 go.sum。）
* chore(deps): 回填 CI 重算的 go.sum
`h1` 由 CI 里真实的 `go mod tidy` 算出后回填——本地没有 Go 工具链，
且 `/go.mod` 的 dirhash 无法自行复现（zip hash 可复现，gomod hash 不可复现），
所以这是唯一权威来源。
重算依赖 `.github` go-ci 里的 `GOPRIVATE=github.com/oagentkit/*` 直连 VCS：
goproxy.cn 仍缓存着版本线重置前的旧内容，走 proxy 会拉回**旧代码**并算出旧 hash，
门禁全绿但依赖是错的。
验证方式就是本 PR 的 `go mod tidy -diff`：若回填有任何偏差，它会照样报差异。
* chore(deps): 回填 CI 重算的 go.sum
`h1` 由 CI 里真实的 `go mod tidy` 算出后回填——本地没有 Go 工具链，
且 `/go.mod` 的 dirhash 无法自行复现（zip hash 可复现，gomod hash 不可复现），
所以这是唯一权威来源。
重算依赖 `.github` go-ci 里的 `GOPRIVATE=github.com/oagentkit/*` 直连 VCS：
goproxy.cn 仍缓存着版本线重置前的旧内容，走 proxy 会拉回**旧代码**并算出旧 hash，
门禁全绿但依赖是错的。
验证方式就是本 PR 的 `go mod tidy -diff`：若回填有任何偏差，它会照样报差异。
* chore(deps): 回填 CI 重算的 go.sum
`h1` 由 CI 里真实的 `go mod tidy` 算出后回填——本地没有 Go 工具链，
且 `/go.mod` 的 dirhash 无法自行复现（zip hash 可复现，gomod hash 不可复现），
所以这是唯一权威来源。
重算依赖 `.github` go-ci 里的 `GOPRIVATE=github.com/oagentkit/*` 直连 VCS：
goproxy.cn 仍缓存着版本线重置前的旧内容，走 proxy 会拉回**旧代码**并算出旧 hash，
门禁全绿但依赖是错的。
验证方式就是本 PR 的 `go mod tidy -diff`：若回填有任何偏差，它会照样报差异。

- 杂务(deps)：跟进最新 tag

oagentkit/textx v0.1.0→v0.2.0
go.sum 取 sumdb 权威 h1，仅就地替换目标 mod 两行。

- 杂务(deps)：oekit 依赖升级

concurrency v0.13.1→v0.16.0、pkg v0.13.0→v0.14.0
姊妹矩阵 oekit 的依赖地板此前被漏掉（上一轮只扫了 oagentkit 内部依赖），本次一并对齐到最新。
升级前已确认 oekit/concurrency v0.13.1..v0.16.0 的 8 个提交全是 feat/docs/chore，
无 breaking/remove 标记。其中 #28 订正了 `WithCooldown(0)` 的语义——
即 llms.txt 里警告过的那个陷阱，升上去正好拿到修正后的行为。
go.sum 的 h1 取自 sumdb，仅就地替换目标 mod 的两行、其余行不动。

- 杂务(deps)：oekit/concurrency v0.12.0 -> v0.13.1

本仓不使用上游 breaker（熔断只在 llm 仓），v0.13.0 的 #25 修复与
v0.13.1 的 #26（RetryableWithHeaders 头值 OWS 修剪）对本仓无直接影响，
本次为矩阵内版本对齐。
门禁：go mod tidy -diff / build / vet /
test -race / gofmt 全绿。

- 杂务(deps)：oekit 三模块升至最新（pkg v0.13.0 / concurrency v0.12.0 / logx v0.7.0）

拿到上游新增能力：
- concurrency #22 RetryableWithHeaders：403 头感知判定（次级限流 / 主限流耗尽）
- concurrency #23 FanInLimit：受限并发扇出
- concurrency #24 WithTokenClock、#21 Lockout（keyed lockout/penalty）
- logx #12 TraceHandler：给裸 slog.Handler 注入 trace_id
破坏性变更核对（均无影响）：
- logx v0.7.0 ParseLogLevel 签名改为 (Level, error) —— 我们未使用该函数
- concurrency v0.9.0 的 Slot.Panicked 废弃 —— rag 已迁到 Slot.Err + errors.As
门禁：go build / go vet / go test -race 全绿。

- 杂务(deps)：oekit/concurrency v0.6.0 -> v0.9.0

拿到上游显式三态状态机（#17）与 singleflight 可观测折叠（#18），并为本次
breaker 替换评估了语义差异。
用到的子包只有 async / singleflight / syncx —— 上游 breaker 包我们并未引用
（runtime/breaker 是本仓自己的三态实现），所以本次升级零 breaking。
关于 breaker 替换的现状：#17 已落地（closed/open/half-open + 代绑定 + 探测时限 +
开放期失败不延长冷却），但两侧的触发语义仍不对齐，已在上游补两个需求：
- #19 冷却期与探测失联时限共用 probeTimeout，无法独立配置
- #20 缺「连续失败 N 次即熔断」（trip）触发
因此本包暂不删除 —— llm.TestResilientBreakerTrips 钉的是 trip 语义，
上游 v0.9.0 只有速率型触发（dropRatio > 0），现在换会让该测试失败。


- CI：接入 org 共享门禁（规则本体在 oagentkit/.github）
