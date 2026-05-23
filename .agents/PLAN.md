# goost 长期规划

本文件是 agent 本地执行计划。公开发布说明继续维护在 `CHANGELOG.md`，迁移说明继续维护在 `MIGRATION.md`；本文件只记录阶段顺序、验证方式和未完成的工程意图。

## 当前契约

`goost` 是一组小而独立的 Go 工具包，当前仍处于 pre-1.0。根 module 必须保持轻量：导入核心包时，不应拉入 demo、benchmark、Gin、gRPC、OpenTelemetry 或其他可选集成依赖。

后续发布工作必须守住这些边界：

- `README.md` 中列出的公开包必须有可编译示例和包级文档。
- 不适合进入根 module 的可选集成依赖，需要保留在 nested modules 中。
- 中文文档需要与英文公开文档保持同义，避免 stale links、缺失 release 引用或迁移建议相互矛盾。
- 本地和 CI 检查统一走 `scripts/check-root.sh` 与 `scripts/check-split-modules.sh`；除非新脚本吸收了新流程，否则不要增加零散的验证路径。

功能面不是冻结状态，但必须受控扩张：新增包或新增公开 API 需要先证明它是可复用的通用工具，而不是某个应用的临时 helper；同时必须在同一轮补齐文档、示例、测试、依赖边界和 changelog。

## 验证阶梯

先使用能覆盖本次改动面的最小 gate；发布前或改动共享脚本时再升级验证范围。

- 仅改 root 包：`./scripts/check-root.sh --quick`
- 仅改一个 nested module：`./scripts/check-split-modules.sh --quick --module <path>`
- 改 root 文档、root 脚本或 README 包表：`./scripts/check-root.sh --quick`
- 改 split-module 文档或示例：运行对应 nested module 的 quick gate
- 改依赖图、CI、发布边界、安全相关逻辑，或做发布前检查：`./scripts/check-root.sh --full` 与 `./scripts/check-split-modules.sh --full`

## 已完成阶段（1-50）

### 阶段 1-5：v0.4.0 发布准备与核心质量

完成 v0.4.0 发布边界对齐（README/CHANGELOG/MIGRATION 中英文一致）、文档与迁移闭环、包一致性整理（每个保留包有示例/文档/测试）、观测语义统一（Stats/Snapshot 统一为不可变快照）、核心并发可靠性强化（race/stress 覆盖 batcher/fanout/pool/ttlmap/keyedmutex）。

### 阶段 6：时间/重试/取消语义收敛

所有包（backoff/batcher/debounce/fanout/pool/ttlmap/httpx/ratelimit）在 context 取消或 Stop/Close 时正确释放 timer 和 goroutine，无泄漏。

### 阶段 7：I/O 与文件系统安全强化

gosec 在 full gate 中保持 0 issues、0 排除。httpx 不记录 query string/body。rotatingwriter 文件权限测试完整（0600/0750）。

### 阶段 8：API 形状与 pre-1.0 兼容性审查

类型命名一致，Option 模式按复杂度选用（functional options/struct config/builder），Error sentinel 使用包前缀，泛型签名统一（[K comparable, V any]/[T any]）。ROADMAP v1.0 audit 全部 "Keep"。

### 阶段 9-10：依赖边界与 CI 自动化

root 与 nested module go.mod/go.sum 保持 tidy。CI 工具版本只声明一次，cache paths 与 go.sum 对齐。

### 阶段 11-12：示例体系与文档体验

每个公开包有最小可编译 example test。大组合示例在 examples/。root README 只做包索引。

### 阶段 13-18：功能扩张评估（Backlog）

所有候选方向（cache/data structure/concurrency/HTTP/config/optional integrations）评估后保留在 backlog，理由：标准库 + 现有包组合已覆盖大部分场景，重依赖不值得独立包。

### 阶段 19-20：v1.0 路线与长期维护

v1.0 readiness list 在 ROADMAP.md，keep/change/remove 决策齐全。maintenance review 流程建立。

### 阶段 21：测试矩阵分层

test matrix 在 TESTING.md，quick/full/stress gate 分层明确。

### 阶段 22：Fuzz 测试

fuzz tests 覆盖 caseconv/env/priorityqueue/lru/ttlmap，不进入 quick gate。

### 阶段 23：性能基线

benchmarks 覆盖 random.String、lru Set/Get、pool Schedule、batcher Load、ratelimit Bucket/Leaky。

### 阶段 24：内存与 goroutine 泄漏防护

所有启动 goroutine/timer 的包有显式 Close/Stop 方法和对应测试。

### 阶段 25：错误模型统一

panic-to-error 路径统一（taskgroup/batcher/pool/shutdown），记录在 API_CONVENTIONS.md。

### 阶段 26：Context 规范化

所有 context.Context API 遵循一致取消语义，记录在 API_CONVENTIONS.md。

### 阶段 27-28：泛型与零值审查

泛型签名统一，不可用零值类型在 doc comment 中标注。

### 阶段 29：私有实现清理

未发现重复 helper、死代码或陈旧兼容残留。

### 阶段 30-31：doc.go 与包命名

root doc.go 与 README 一致。包命名全部 "Keep"。

### 阶段 32-35：自动化校验与模块治理

smoke tests 校验 package table/README/example 对应。split-module gate 覆盖 examples。lru/benchmark 独立模块。

### 阶段 36-38：安全、日志与集成审计

gosec 0 issues。zapctx/slogctx API 对齐。zapctxgin/zapctxgrpc 有完整 README/example/payload test。

### 阶段 39-42：发布脚本、Go 版本、跨平台、压力测试

check-release.sh 存在。go.mod 1.25.10 与 CI 一致。shutdown 有 `//go:build !windows` 测试。stress tests 在独立文件。

### 阶段 43-45：贡献入口、API 弃用、组合验证

issue/PR 模板存在。无 deprecated API（pre-1.0）。examples/ 有组合示例。

### 阶段 46-48：术语、迁移、脚本质量

术语在 API_CONVENTIONS.md 统一。testdata/migration/ 有迁移 fixture。check-scripts.sh 验证脚本质量。

### 阶段 49-50：项目定位复盘与路线图复审

PROJECT_POLICY.md 记录范围边界。ROADMAP.md 当前且与代码同步。

---

## 阶段 51-70：下一轮规划

### 阶段 51 - v0.4.0 发布执行

目标：正式发布 v0.4.0，将 Unreleased 线收束为 tag。

任务：

- 运行 `./scripts/check-release.sh` 并记录通过结果。
- 在 CHANGELOG.md 中将 `[Unreleased]` 替换为 `[v0.4.0]` 及日期，同步中文 CHANGELOG。
- 在 MIGRATION.md 中确认 v0.4.0 迁移说明完整。
- 更新 README.md 和 README.zh.md 中的版本引用。
- 创建 git tag `v0.4.0` 并推送。
- 在 GitHub Releases 创建发布说明。

完成标准：

- `v0.4.0` tag 在远程仓库可查。
- CHANGELOG、MIGRATION、README 中英文一致。
- 发布说明包含主要变更摘要。

### 阶段 52 - godoc 渲染质量审查

目标：确保每个包在 pkg.go.dev 上渲染出清晰、完整的 API 文档。

任务：

- 审查所有包的 package doc comment：是否有一句话概述、核心类型列表、使用示例引用。
- 检查 exported types/funcs/methods 的 doc comment 完整性：参数、返回值、行为边界。
- 确认 `doc.go` 与主文件 package comment 不重复；对注释较长的包（如 httpx、pool、batcher），考虑拆出独立 `doc.go`。
- 为 `clock.Clock` 接口和 `httpx.Limiter` 接口的每个方法补齐 doc comment。
- 检查 `Example*` 函数的 Output 注释是否与实际输出一致，确保 godoc 能正确渲染。

完成标准：

- 每个 exported symbol 都有 doc comment。
- package doc comment 覆盖一句话概述和核心用法。
- `go doc ./...` 输出无空白项。

### 阶段 53 - 测试覆盖率门槛建立

目标：为 root module 建立可追踪的覆盖率基线，防止覆盖率无意退化。

任务：

- 使用 `go test -coverprofile` 生成当前覆盖率报告，按包记录基线。
- 识别覆盖率低于 80% 的包，评估是否需要补充测试。
- 对低覆盖率的包，优先补充错误路径、边界输入和关闭/取消路径测试。
- 不设硬性门槛阻断 CI，但在 ROADMAP 中记录基线值，供后续跟踪。
- 考虑在 `scripts/check-root.sh --full` 中输出覆盖率摘要。

完成标准：

- 每个包的覆盖率基线已记录。
- 低于 80% 的包有明确的补测试计划或书面理由（如代码路径不可达）。
- full gate 输出包含覆盖率摘要。

### 阶段 54 - 错误链与 Sentinel 审计

目标：统一所有包的错误处理模式，确保错误链在 `errors.Is`/`errors.As` 下可遍历。

任务：

- 盘点所有 exported sentinel errors（`ErrNotFound`、`ErrOpen`、`ErrPoolClosed` 等），确认命名一致（`Err` 前缀 + 包名前缀消息）。
- 检查 `fmt.Errorf("pkg: %w", err)` 的 wrap 模式是否一致，特别是 batcher、pool、httpx、taskgroup 中。
- 审查 `errors.Join` 和 `multierr` 模式是否在 taskgroup/results 和 fanout 中行为一致。
- 确认所有公开错误都能被 `errors.Is` 匹配，而不是只能字符串比较。
- 为 batcher 的 `ErrNotFound` 评估是否需要 `errors.Is` 语义（当前是 sentinel）。

完成标准：

- 错误命名和 wrap 模式在 API_CONVENTIONS.md 中有统一约定。
- 所有公开错误都有对应的 `errors.Is` 测试。

### 阶段 55 - Consumer Contract Tests

目标：为公开 API 建立 consumer-side 测试，确保 import path 和 API 签名不会被意外破坏。

任务：

- 在 `dependency_smoke_test.go` 中增加 API 签名验证：确认 exported types 的字段名和类型未改变。
- 为每个包的核心类型增加 compile-time 接口满足检查（`var _ io.Writer = (*RotatingWriter)(nil)` 模式）。
- 为 optional integration modules 增加独立 consumer test，验证 `go get` 路径正确。
- 评估是否需要单独的 `testdata/consumer/` 目录存放外部消费者视角的编译测试。

完成标准：

- 核心类型有 compile-time 接口检查。
- dependency smoke test 覆盖 API 表面稳定性。
- nested module 的 import path 有编译验证。

### 阶段 56 - 依赖更新自动化

目标：建立外部依赖的自动更新和审查流程，减少手动跟踪负担。

任务：

- 评估是否启用 Dependabot 或 Renovate：对 Go modules、GitHub Actions 版本、CI 工具版本。
- 配置更新策略：patch 自动合并、minor 需审查、major 需手动。
- 确认 `scripts/install-ci-tools.sh` 中的工具版本与自动化配置一致。
- 为 nested modules 建立与 root 相同的更新策略。

完成标准：

- 依赖更新有自动化管道。
- CI 工具版本更新不需要手动编辑多处。
- 更新不会绕过 full gate 验证。

### 阶段 57 - Benchmark Regression CI Gate

目标：在 CI 中建立性能退化检测，防止关键路径无意变慢。

任务：

- 选择 3-5 个性能敏感包（lru、pool、batcher、ratelimit、random）的核心 benchmark。
- 在 CI 中增加一个可选的 benchmark job，记录每次运行的 `ns/op` 和 `allocs/op`。
- 不阻断普通 PR，仅在结果中展示对比，供审查者判断。
- 为 benchmark 结果建立存储方案（JSON artifact 或 git notes）。
- 在 TESTING.md 中记录 benchmark CI 使用方式。

完成标准：

- CI 可运行 benchmark 并输出对比。
- 性能退化能在 PR review 中被观察到。
- 不增加 quick gate 的运行时间。

### 阶段 58 - Examples 冒烟自动化

目标：让 `examples/` 中的 6 个可执行程序在 CI 中被自动编译和冒烟运行。

任务：

- 评估每个 example 是否能在 CI 中安全运行（httpserver 需要端口、eventbus 可能需要时间）。
- 为可安全运行的 examples 增加 `go build` + `go run`（带 timeout）验证。
- 在 split-module gate 中增加 examples 编译检查。
- 确认 examples 的 README 与实际目录结构一致。

完成标准：

- 每个 example 都能在 CI 中编译。
- 可安全运行的 example 在 CI 中实际执行并退出。
- 不引入 flaky 或外部依赖。

### 阶段 59 - CHANGELOG 自动化辅助

目标：降低 CHANGELOG 维护成本，减少跨文件同步遗漏。

任务：

- 评估 conventional commits 或类似约定对 PR 提交信息的适用性。
- 如果采纳，增加 PR 模板中的变更分类指引。
- 在 `scripts/check-release.sh` 中增加 CHANGELOG 格式校验（Unreleased heading 存在、分区格式正确）。
- 增加中英文 CHANGELOG 同步检查：确认两边 release heading 一致。
- 不强制自动生成，但提供结构化辅助。

完成标准：

- PR 模板引导贡献者填写 CHANGELOG 条目。
- release gate 校验 CHANGELOG 格式。
- 中英文 CHANGELOG 结构不漂移。

### 阶段 60 - Go 版本策略细化

目标：明确最低支持 Go 版本、测试策略和升级节奏。

任务：

- 确认 `go.mod` 中的 `go` 指令与 CI `GO_VERSION` 的语义关系：`go` 指令是最低版本，CI 用最新稳定版。
- 评估是否增加 `oldest-supported` Go 版本的 CI matrix job。
- 在 README 和 CONTRIBUTING 中明确说明最低 Go 版本要求。
- 为 Go 新版本（如 1.26+）建立升级 checklist：更新 go.mod、CI、install scripts、CODEOWNERS。

完成标准：

- 用户能从 README 明确知道最低 Go 版本。
- CI 覆盖最低版本和最新版本。
- Go 版本升级有流程文档。

### 阶段 61 - 可观测性 Hook 一致性验证

目标：确保所有包的 callback/hook 签名、调用时机和 panic 处理一致。

任务：

- 盘点所有 hook/callback API：httpx OnRetry/OnGiveUp/Logger、pool WithPanicHandler、circuitbreaker OnStateChange、batcher OnFlush、fanout OnDrop。
- 统一 hook 的调用时机（before/after）、是否可能 block 主路径、panic recovery 策略。
- 为每个 hook 补测试：验证在 callback panic 时不影响主路径。
- 在 API_CONVENTIONS.md 的 Observability 节中补齐 hook 约定。

完成标准：

- 所有 hook 有文档说明调用时机和 panic 策略。
- hook panic 不影响主路径（有测试覆盖）。
- API_CONVENTIONS.md 记录统一约定。

### 阶段 62 - 表驱动测试模式统一

目标：让测试风格在包间保持一致，降低阅读和维护成本。

任务：

- 审查所有包的测试文件，识别非表驱动风格的测试（尤其是长 if/else 链）。
- 对可自然转为 table-driven 的测试进行重构，不改变测试覆盖面。
- 统一 test helper 命名：`testXxx` vs `newTestXxx` vs `fakeXxx`。
- 统一 error 检查模式：`t.Fatal` vs `require.NoError`（当前项目用 `t.Fatal`/`t.Errorf`，保持一致）。
- 不强制所有测试变为表驱动；只改明显收益的。

完成标准：

- 非表驱动测试有合理理由（在代码注释中说明）。
- test helper 命名一致。
- 错误检查模式一致。

### 阶段 63 - 结构化日志测试辅助

目标：为使用 zap/slog 的包提供可测试的日志验证方式。

任务：

- 审查 zapctx 和 slogctx 的测试：是否验证了字段名和字段值。
- 为 httpx 的 Logger callback 增加 structured output 测试（验证不泄露敏感字段）。
- 评估是否需要在 test helper 中暴露 `slogtest` 或 zap test logger 的封装。
- 确认 examples 中的日志输出是可预测的（无时间戳、无随机 ID）。

完成标准：

- 日志输出测试覆盖字段名和字段值。
- 无敏感字段泄漏到日志输出（有测试断言）。
- 日志测试不受时间戳或随机值影响。

### 阶段 64 - 跨包组合集成测试

目标：为核心包组合路径建立独立的集成测试，超越 examples/ 中的 demo 级验证。

任务：

- 设计 5-8 个组合场景：backoff+circuitbreaker+httpx、pool+taskgroup+shutdown、lru+ttlmap 缓存策略、batcher+fanout 事件分发、ratelimit+httpx 限流。
- 组合测试放在 `integration/` 目录或 `_test.go` 文件中，不引入新依赖。
- 每个组合测试验证包间接口兼容性、错误传播、取消传播和关闭顺序。
- 不变成长篇应用，每个测试保持 50 行以内。

完成标准：

- 核心组合路径有独立测试。
- 组合测试覆盖取消传播和关闭顺序。
- 不引入外部依赖。

### 阶段 65 - API 选项验证一致性

目标：确保所有包的 constructor/builder 在传入无效配置时行为一致。

任务：

- 盘点所有 constructor/builder 的参数验证：pool.NewPool、batcher.New/Build、lru.New+Builder、fanout.New+Builder、ratelimit.NewBucket/NewLeaky、circuitbreaker.New、debounce.New。
- 统一 panic-on-invalid-config（用于编程错误）vs return-error（用于运行时参数）的边界。
- 确认所有 builder 的零值/nil 处理一致（忽略 nil option vs panic）。
- 为每个 constructor 补无效输入测试。

完成标准：

- 无效配置在 constructor 中被立即检测。
- panic vs error 的边界在 API_CONVENTIONS.md 中有约定。
- 每个 constructor 有无效输入测试覆盖。

### 阶段 66 - 文档链接完整性校验

目标：确保所有 README 和文档中的交叉引用链接有效。

任务：

- 在 `scripts/check-release.sh` 或独立脚本中增加 Markdown 链接校验。
- 检查范围：README 中的相对路径链接、CHANGELOG 中的版本链接、MIGRATION 中的包引用。
- 检查中英文文档的交叉链接不指向错误语言版本。
- 不检查外部 URL（可能有网络问题），只检查仓库内相对路径。
- 对 godoc 链接验证格式正确性。

完成标准：

- release gate 包含文档链接校验。
- 中英文文档链接不互相引用错误。
- 断链在 CI 中被自动捕获。

### 阶段 67 - 并发包 WaitGroup/Once 模式审计

目标：确保所有使用 sync.WaitGroup、sync.Once、sync.Map 的包遵循正确模式。

任务：

- 审查 WaitGroup 使用：Add/Done 配对、Add 在 Wait 之前、不可复用。
- 审查 sync.Once 使用：确认只用于初始化，不用于每次请求的 lazy init。
- 审查 sync.Map 使用：评估是否可用 `defaultmap` 或普通 map+mutex 替代（减少 API 表面积）。
- 为发现的模式问题补测试。

完成标准：

- WaitGroup 使用模式正确（有测试验证）。
- sync.Once 不被误用。
- sync.Map 使用有明确理由。

### 阶段 68 - 发布后 hotfix 流程建立

目标：为已发布版本建立快速修复流程，减少紧急修复的响应时间。

任务：

- 在 CONTRIBUTING.md 中增加 hotfix 流程说明。
- 明确 hotfix 需要的最小验证：quick gate + 受影响包的 full gate。
- 确认 CHANGELOG hotfix 格式：在现有 release 下新增条目 vs 新建 patch release。
- 为 hotfix 建立快速 review 路径（不需要 full release gate，但需要相关包 gate）。

完成标准：

- hotfix 有文档化流程。
- hotfix 不需要跑完 full release gate 就能发布 patch 版本。
- CHANGELOG 格式支持 patch release。

### 阶段 69 - Go 1.26+ 兼容性准备

目标：为下一个 Go 版本建立前瞻性兼容检查。

任务：

- 跟踪 Go 1.26 release notes 中的语言和标准库变更。
- 检查是否有被废弃的标准库 API 被项目使用。
- 评估新标准库能力是否可以简化现有包（如新的 sync primitives、testing helpers）。
- 在 CI matrix 中增加 Go 1.26 rc/beta 测试（允许失败）。
- 更新 go.mod 工具链指令。

完成标准：

- Go 1.26 兼容性有明确评估。
- 被废弃的 API 有迁移计划。
- 新标准库能力有评估结论。

### 阶段 70 - 长期路线图复审与 v0.5.0 规划

目标：基于阶段 51-69 的执行结果，复审整体路线并规划 v0.5.0 内容。

任务：

- 汇总阶段 51-69 的完成情况，更新 ROADMAP.md。
- 评估 v0.4.0 发布后的用户反馈，调整 backlog 优先级。
- 为 v0.5.0 确定范围：是否有功能扩张（阶段 13-18 backlog 中是否有提升项）。
- 清理 PLAN.md 中已完成的阶段摘要，保持文件可操作。
- 为下一轮 10-20 个执行切片排序。

完成标准：

- ROADMAP.md 反映 v0.4.0 后的真实状态。
- v0.5.0 范围有明确界定。
- 下一轮执行切片有 surface/artifact/validation。
- PLAN.md 持续保持可操作性。

## Backlog Ideas

- 只有当包存在真实性能声明或风险时才考虑 benchmark，例如 cache eviction、batch fan-in、retry overhead 或 queueing。
- 可考虑为 optional modules 增加 consumer contract tests，确认 import paths 稳定，同时依赖仍留在 root module 之外。
- 如果 v0.4.0 发布前仍需要大量手工跨文件核对，可考虑增加 release checklist script。
- 评估是否需要 Go module workspace（go.work）来简化本地开发中的跨模块修改。
- 评估是否需要 OpenTelemetry tracing 集成作为 nested module（基于用户反馈决定）。
- 考虑增加 `staticcheck` 的自定义检查规则，覆盖项目特有的 API 约定。

## 审查记录（阶段 51-70 执行总结）

### 阶段 51 - v0.4.0 发布执行

`./scripts/check-release.sh` 已通过。`v0.4.0` 本地 tag、远程 tag 和 GitHub Release 已创建，均指向 `89fc4e9`（`chore: prepare v0.4.0 release with quality infrastructure (phases 51-70)`）。GitHub Release：`https://github.com/liguangsheng/goost/releases/tag/v0.4.0`，发布时间 `2026-05-23T18:07:07Z`。

### 阶段 52 - godoc 渲染质量审查

补齐 caseconv（snake/kebab/camel）、rotatingwriter（NewRotatingWriter/NewDailyRotater）、zapctxgin（Middleware）、zapctxgrpc（UnaryServerInterceptor）的 exported symbol doc comments。`go vet ./...` 和 `go doc ./...` 无空白项。

### 阶段 53 - 测试覆盖率门槛建立

总体覆盖率 91.8%。当前没有 root 包低于 80%。覆盖率基线记录在 TESTING.md 和 TESTING.zh.md。

### 阶段 54 - 错误链与 Sentinel 审计

为 backoff（PermanentError）、batcher（ErrNotFound）、circuitbreaker（ErrOpen）、ratelimit（ErrLimitExceeded）添加 errors.Is/errors.As 兼容性测试。

### 阶段 55 - Consumer Contract Tests

添加编译时接口检查：rotatingwriter.RotatingWriter 实现 io.Writer，DailyRotater/SizeRotater 实现 Rotater，clock.Mock 实现 Clock，ratelimit.Bucket 实现 httpx.Limiter。

### 阶段 56 - 依赖更新自动化

创建 .github/dependabot.yml，配置 root module、所有 nested modules 和 GitHub Actions 的自动更新。

### 阶段 57 - Benchmark Regression CI

在 ci.yml 中增加可选的 benchmark job（PR 触发、continue-on-error），输出 benchmark 结果到 GitHub Step Summary。

### 阶段 58 - Examples 冒烟自动化

在 ci.yml 中增加 examples-smoke job，遍历 examples/*/ 编译所有可执行程序。

### 阶段 59 - CHANGELOG 自动化辅助

在 check-release.sh 中增加 CHANGELOG 格式校验（release heading 和 section heading 存在性检查）。

### 阶段 60 - Go 版本策略细化

在 CONTRIBUTING.md 中增加 Go Version Policy 节，说明 go.mod 下限和 CI 上限的关系。

### 阶段 61-67 - 代码质量测试

Phase 61（hook panic safety）补齐为当前代码事实：`httpx` 的 OnRetry/OnGiveUp、`circuitbreaker` 的 OnStateChange、`pool` 的 PanicHandler 都同步执行但会 recover callback panic，不改变宿主操作结果；API_CONVENTIONS 中英文同步记录 hook 约定。新增测试覆盖 callback panic 后主流程继续。Phase 65（constructor validation）和 Phase 67（sync 审计）已审查确认：构造函数验证模式一致（pool 返回 error，ratelimit panic）；sync 原语使用正确（WaitGroup Add/Done 配对，Once 用于初始化）。验证：`go test ./httpx ./circuitbreaker ./pool`，`./scripts/check-root.sh --quick`。

### 阶段 62-63 - 表驱动测试和日志测试

完成审查并补齐可验证证据。Phase 62：扫描 85 个 `_test.go`，未发现值得机械改写的长 if/else 测试；保留并发、生命周期、select、fuzz 和 smoke checks 的直接场景写法，在 TESTING.md / TESTING.zh.md 记录实际约定（纯输入输出优先 table-driven，package-level expectations 用 assert，前置条件用 require，smoke/fuzz/select 用 t.Fatal/t.Errorf，helper 名称直接描述角色）。Phase 63：`httpx` 增加结构化 `slog.Handler` 测试，断言 method/scheme/host/path/status/attempts 字段值且不泄露 query/body secret；`zapctx` 增加 zap observer 字段值和 sampled 行为测试；`slogctx` 已有字段输出测试。验证：`go test ./httpx ./zapctx ./slogctx`。

### 阶段 64 - 跨包组合集成测试

添加 6 个集成测试：backoff+circuitbreaker、pool+shutdown、ratelimit+httpx、lru+ttlmap、batcher+fanout、taskgroup+errors。

### 阶段 66 - 文档链接完整性校验

创建 scripts/check-doc-links.sh，集成到 check-release.sh。所有 markdown 相对链接验证通过。

### 阶段 68 - 发布后 hotfix 流程

在 CONTRIBUTING.md 中增加 Hotfix Process 节，明确 patch release 的最小验证要求。

### 阶段 69 - Go 1.26+ 兼容性分析

Go 1.26 已发布（官方 release notes：`https://go.dev/doc/go1.26`），当前最新下载元数据为 go1.26.3（`https://go.dev/dl/?mode=json`）。保留 `go.mod` 的 `go 1.25.10` 作为当前主支持/最低本地工具链，CI 新增允许失败的 `go-next-root-smoke` job，用 Go 1.26.3 跑 `./scripts/check-root.sh --quick`，用于提前发现语言、标准库、timer/channel、平台支持和 toolchain 兼容性漂移。README/README.zh 和 ROADMAP/ROADMAP.zh 已同步记录 Go 1.26.3 compatibility probe。

### 阶段 70 - 长期路线图复审

ROADMAP.md slices 32-43 已执行。PLAN.md 审查记录更新。`./scripts/check-release.sh` 和 `./scripts/check-root.sh --quick` 通过。
