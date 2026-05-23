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

## 阶段 1 - 稳定 v0.4.0 发布边界

目标：在移除低价值公开包并完成 split-module 清理后，把当前 Unreleased 线收束成一致的 v0.4.0 候选状态。

任务：

- 对齐 `README.md`、`README.zh.md`、`CHANGELOG*.md` 和 `MIGRATION*.md`，确保每个移除包、nested module 和行为变化在中英文中表达一致。
- 每次改 README 包表后，重新跑 dependency smoke coverage，确认核心包仍不会导入可选 Gin、gRPC 或 tracing 依赖。
- 确认 optional module split 后，root 与 nested `go.mod` / `go.sum` 都保持 tidy。
- 对仍只写在 release notes、但缺少断言的行为补测试或收紧现有测试。
- 打 tag 前运行 root 与 split-module full gates，并在发布工作流或最终交接中记录准确命令结果。

完成标准：

- 本地 full gates 通过。
- 公开文档和迁移说明在中英文中一致。
- `CHANGELOG.md` 的信息足够支撑 v0.4.0 tag，不依赖本文件解释。

## 阶段 2 - 发布前文档与迁移闭环

目标：让 v0.4.0 的公开叙事完整、可验证，并能让使用者从 v0.3.x 顺利迁移。

任务：

- 逐项核对 `CHANGELOG*.md` 与 `MIGRATION*.md`，确认每个 breaking/removal/change 都有迁移建议。
- 检查 README 包表、optional modules 表和实际目录、`go.mod` 拆分状态一致。
- 确认中文文档不链接英文 release docs，英文文档不依赖中文私有说明。
- 为 release boundary 增加必要的 smoke tests，而不是只靠人工审阅。

完成标准：

- 用户只读公开文档即可理解 v0.4.0 的变化。
- release 文档不需要引用 `.agents/PLAN.md` 才能解释清楚。

## 阶段 3 - 包一致性整理

目标：让每个保留下来的包都像被认真维护的公开 API，具备一致的 API、示例、测试和文档质量。

任务：

- 审计 README 中列出的每个包：package comment、英文 README、中文 README、可编译示例、常规测试，以及并发核心包所需的 race/stress 覆盖。
- 统一暴露 `Stats` 或 `Snapshot` 的包的观测语义，明确这些值是瞬时值、单调值，还是在锁内派生出的只读视图。
- 检查 `batcher`、`debounce`、`ttlmap`、`httpx`、`pool` 和集成模块的 option builder 与 nil-handling 语义是否一致。
- 示例保持短小且可执行；只有需要端到端形态的较大程序才放入 `examples/`。
- 审计中发现边界问题时，优先补有针对性的回归测试，避免顺手大重写。

完成标准：

- README 包表与真实公开包集合一致。
- 每个列出的包都有可编译示例和最新的中英文文档。
- 保留下来的不一致行为都有明确理由，并在文档中说明。

## 阶段 4 - 观测语义统一

目标：让所有 `Stats`、`Snapshot`、callback 和 logging hook 的语义一致，避免每个包各说各话。

任务：

- 盘点 `batcher`、`fanout`、`pool`、`ratelimit`、`ttlmap`、`lru`、`circuitbreaker`、`httpx` 的观测出口。
- 明确每个字段是当前值、累计值、配置值还是派生值。
- 在 README 和 doc comment 中说明线程安全、时间点和锁语义。
- 为容易误用的观测字段补测试，尤其是关闭后、并发中、错误路径上的值。

完成标准：

- 观测 API 的命名和文档能形成统一模式。
- 新增观测能力时有现成参照，不再重新发明风格。

## 阶段 5 - 核心可靠性与并发强化

目标：提高涉及 goroutine、timer、I/O 或 retry 的包的可信度，同时避免无谓扩大公开 API。

任务：

- 为 `batcher`、`fanout`、`pool`、`taskgroup`、`ttlmap`、`ratelimit` 和 `rotatingwriter` 增强 race-sensitive 回归测试，覆盖 cancellation、shutdown、queue pressure、expiration 和 cleanup 路径。
- 检查 fake-clock 测试是否覆盖边界时间点，避免依赖 sleep 造成 flaky。
- 每次改 `httpx` transport 或 observability 逻辑后，重新审查 retry 与 body replay 路径。
- `rotatingwriter` 的文件权限和路径处理测试要保持明确；不要为了让示例通过而削弱安全扫描覆盖。
- 只有多个包确实需要同一行为时，才考虑公开抽象；否则优先使用小的内部 helper。

完成标准：

- 并发核心包具备可实际运行的 `go test -race` 覆盖。
- full gates 通过，并且没有压掉 `gosec`、`govulncheck`、`staticcheck` 或 lint 发现。

## 阶段 6 - 时间、重试与取消语义收敛

目标：让涉及时间、重试、限流和取消的包在语义上保持可预测。

任务：

- 统一 `clock`、`backoff`、`ratelimit`、`debounce`、`httpx` 中 context cancellation 和 timer cleanup 的预期行为。
- 检查 fake-clock 测试是否覆盖边界推进、零值 duration、负值 duration 和并发唤醒。
- 确认 retry delay、limiter wait、debounce quiet window 都不会在 context 取消后泄漏 goroutine 或 timer。
- 把跨包一致的边界写入 README 或 example，而不是只体现在测试里。

完成标准：

- 时间相关包在边界输入下行为稳定。
- 用户能从文档中判断取消和超时会如何传播。

## 阶段 7 - I/O 与文件系统安全强化

目标：把涉及文件、HTTP、日志 payload 的包做成默认安全、可审计的工具。

任务：

- 继续强化 `rotatingwriter` 的权限、路径、备份清理和 gzip 行为测试。
- 审查 `httpx` 日志输出，确认不会记录 query string、body 或高风险 header。
- 对 examples 中的 HTTP server、文件路径和临时目录使用保持安全扫描友好。
- 安全扫描误报必须用代码或注释解释，不能直接降低 gate 强度。

完成标准：

- `gosec ./...` 能在 full gate 中保持无排除运行。
- 文件与日志相关行为在文档和测试中都有明确边界。

## 阶段 8 - API 形状与 pre-1.0 兼容性审查

目标：利用 pre-1.0 的窗口，把偶然复杂度清掉，再让公开表面逐步稳定。

任务：

- 审查 exported names、option names、stats structs 和 error values，找出一致性问题和明显迁移风险。
- 识别过窄、过度 application-specific，或已被标准库充分覆盖的包，在未来 v1.0 之前决定保留、合并或移除。
- 只要 exported symbol 发生变化或迁移，就把迁移说明写具体。
- 避免为了省几行代码让小工具包之间互相依赖；只有耦合收益明确时才接受跨包依赖。
- 对兼容性敏感的变更使用 consumer-style tests，尤其是 nested modules 和集成 import paths。

完成标准：

- 为剩余破坏性决策留下简短的 v1.0 readiness list。
- 每个 breaking change 都有测试、迁移说明和 changelog 覆盖。

## 阶段 9 - 依赖边界与 module 拆分治理

目标：长期保持“按需依赖”的核心定位，不让 root module 被可选功能拖重。

任务：

- 为 root 包、nested modules、examples、benchmarks 建立清晰的归属规则。
- 每次新增外部依赖时，判断它属于 root、nested module 还是 example-only。
- 扩展 dependency smoke tests，覆盖 optional integrations 不回流 root module 的约束。
- 保持 CI cache dependency paths、split-module gates 和实际 `go.sum` 文件一致。

完成标准：

- `go get github.com/liguangsheng/goost` 的依赖图仍然可解释、可控。
- 新 nested module 的加入不会让 CI 或本地 gate 变得隐蔽。

## 阶段 10 - 工具链、CI 与发布自动化

目标：降低维护成本，让 CI 失败能够直接指向被破坏的边界。

任务：

- Go 版本和工具版本尽量只在 CI 中声明一次，并由安装脚本复用。
- 每次新增、移除或重命名 nested module 时，同步维护 cache dependency paths。
- root 与 split-module 脚本是唯一受支持的本地 gate；脚本行为变化时同步更新 help text。
- 只有某类错误反复出现，或手动排查成本很高时，才新增脚本检查。
- 定期核对 GitHub Actions major versions、Node runtime 要求和 Codecov 行为是否仍匹配当前 runner 环境。

完成标准：

- CI 与文档中声明的本地 full gates 对齐。
- 新增或移除 nested module 时，有清晰 checklist 和脚本支持。

## 阶段 11 - 示例体系与使用路径整理

目标：让示例覆盖真实使用路径，同时保持依赖轻、输出稳定、容易阅读。

任务：

- 将 package-level examples 保持为最小可编译用法。
- 将端到端 runnable examples 留在 `examples/`，并明确它们覆盖的组合场景。
- 避免示例引入随机、网络、真实时间 sleep 或安全扫描不友好的写法。
- 为常见组合路径补示例，例如 retry + logging、pool + shutdown、cache + ttl。

完成标准：

- 每个公开包都有足够小的 example test。
- 大示例展示组合能力，但不影响 root module 的依赖边界。

## 阶段 12 - 文档与采用体验打磨

目标：让使用者可以按包快速评估这个库，而不是把 root README 写成框架手册。

任务：

- root README 只承担包索引、安装说明、稳定性说明和开发 gate 入口。
- 包级 API 细节放在 package README 和 example 中，不塞进 root README。
- 英文文档变化时同步中文文档，尤其是 release、migration 和依赖边界相关文字。
- 示例保持真实但克制依赖；大型 demo 留在 split modules 或 `examples/`。
- 新增包时，在同一切片内补齐文档、example tests、dependency smoke coverage 和 changelog entries。

完成标准：

- 新用户不需要读私有计划文件，也能找到合适的包并运行示例。
- 公开文档不提 `.agents` 或 agent-specific workflow。

## 阶段 13 - 受控功能扩张总则

目标：在现有边界稳定后继续扩展功能面，但只增加真正适合作为通用库沉淀的能力。

候选方向：

- 补齐现有包自然延伸出的能力，例如 `httpx` 的更完整 retry/observability hooks、`rotatingwriter` 的更清晰 retention 策略、`pool` / `taskgroup` 的运行态观测辅助。
- 增加与当前主题一致的新小包，例如轻量 cache primitives、request/response helper、concurrency coordination helper、配置解析辅助等，但必须先确认标准库或现有包不能自然覆盖。
- 为 optional integrations 增加新 nested modules，而不是把重依赖塞回 root module。
- 增加示例程序时优先服务真实使用路径，不为了展示而扩大依赖面。

准入标准：

- API 能用一句话说明清楚，且至少有两个合理使用场景。
- 不引入不必要的 root module 依赖；重依赖默认进入 nested module。
- 同一切片内补齐 README、README.zh.md、example test、普通测试、dependency smoke coverage 和 changelog。
- 如果新增能力会扩大维护负担，先写在 backlog，等现有 release boundary 稳定后再实现。

完成标准：

- 每个新增功能都有明确归属：扩展现有包、增加新 root 包，或增加 nested module。
- 新功能不会破坏“按需依赖”的核心定位。
- full gates 通过，公开文档能独立解释新增能力。

## 阶段 14 - 缓存与数据结构扩张

目标：在现有 `lru`、`ttlmap`、`defaultmap`、`priorityqueue` 基础上补齐常用但仍轻量的数据结构能力。

候选方向：

- 审查是否需要通用 cache interface、bounded map、single-key loading cache 或带 TTL 的 LRU 组合能力。
- 评估 `defaultmap` 与 `ttlmap` 是否有可复用的 lazy construction / expiration 模式。
- 仅在标准库、现有包和小型组合无法覆盖时新增 root 包。

完成标准：

- 新增数据结构必须是泛型、低依赖、易测试的 root 包候选。
- 不为单一业务场景增加专用容器。

## 阶段 15 - 并发协调能力扩张

目标：围绕 `batcher`、`fanout`、`pool`、`taskgroup`、`keyedmutex` 增强通用并发协调能力。

候选方向：

- 评估是否需要 semaphore、once-per-key、debounced worker、bounded fan-in/fan-out 等小工具。
- 为现有包补齐 context-aware variants，而不是重复创建相近包。
- 每个新增能力都必须有 race tests 和 cancellation tests。

完成标准：

- 并发工具之间分工清晰，没有名称相近但语义重叠的 API。
- 新增并发功能在 race gate 下稳定。

## 阶段 16 - HTTP 与服务端辅助扩张

目标：让 `httpx` 和相关 helper 能覆盖常见服务调用场景，同时不变成完整 web framework。

候选方向：

- 扩展 `httpx` 的 retry/give-up/metrics hooks、request classification、response drain 策略和测试覆盖。
- 评估是否需要小型 request/response helpers，例如 typed JSON decode、status classifier、safe header utilities。
- 重依赖 web framework 的能力放入 nested modules，不进入 root。

完成标准：

- `httpx` 仍是标准库 `*http.Client` 的薄增强，而不是封闭客户端框架。
- 新增 HTTP helper 不扩大敏感日志面。

## 阶段 17 - 配置、环境与启动关闭扩张

目标：围绕 `env`、`shutdown`、`clock` 等包补齐服务启动阶段常用的轻量能力。

候选方向：

- 评估 `env` 是否需要默认值、required 字段、duration/slice 类型和自定义 decoder 的更完整测试或 API。
- 评估 `shutdown` 是否需要 readiness/drain 状态、hook 分组、错误聚合策略。
- 保持 API 小而显式，不引入配置框架或生命周期框架。

完成标准：

- 服务启动/关闭相关能力能独立使用，也能与 `httpx`、`pool`、`taskgroup` 组合。
- 文档清楚说明错误处理和执行顺序。

## 阶段 18 - Optional integrations 生态扩展

目标：为常用外部生态提供可选集成，但所有重依赖都必须留在 nested modules。

候选方向：

- 在 `zapctx/zapctxgin`、`zapctx/zapctxgrpc` 模式成熟后，评估是否需要 slog/gin、slog/grpc 或其他日志集成 nested modules。
- 集成包必须只做桥接，不把外部框架概念反向带入 root API。
- 每个 nested module 都要有独立 README、example test、go.mod、CI cache path 和 split gate 覆盖。

完成标准：

- 用户按需导入集成 module，root module 依赖图不变化。
- optional integrations 的文档明确说明 import path、module path 和适用场景。

## 阶段 19 - v1.0 稳定化路线

目标：当功能面和质量门槛足够稳定后，开始收敛 v1.0 API。

任务：

- 整理所有 remaining breaking decisions，按包列出 keep/change/remove。
- 将 pre-1.0 期间的迁移建议压缩成 v1.0 前最终迁移指南。
- 明确 v1.0 后的兼容性承诺、deprecation policy 和 release cadence。
- 冻结 root package set 前做一次 full dependency、docs、examples、race、安全审计。

完成标准：

- v1.0 readiness list 中没有模糊项。
- 公开文档能说明 v1.0 后用户可以依赖哪些稳定承诺。

## 阶段 20 - 长期维护与淘汰机制

目标：让项目在 v1.0 后仍能健康演进，既能新增能力，也能淘汰低价值表面。

任务：

- 建立周期性审计：依赖、CI、文档链接、examples、security scans、Go 版本。
- 为低使用、低价值或标准库已覆盖的 API 制定 deprecation 流程。
- 保持 changelog、migration、README 和 tests 同步更新，不让维护知识只存在于私有计划中。
- 每隔一段时间重新评估功能扩张 backlog，删除不再成立的想法。

完成标准：

- 新增、变更、弃用都有一致流程。
- 项目长期保持轻量、可维护、按需依赖的定位。

## 阶段 21 - 测试矩阵分层治理

目标：把快速反馈、并发验证、安全扫描和发布前验证分成清晰层级，避免所有工作都只能跑 full gate。

任务：

- 明确 root quick、root full、split quick、split full、package targeted test 的适用场景。
- 为高风险包保留可单独运行的 race test 命令，降低定位成本。
- 检查 CI 是否能在失败时快速指出 root、nested module、lint、security 或 docs boundary。

完成标准：

- 开发者能根据改动面选择最小有效验证。
- full gate 仍覆盖发布前必须承担的全部风险。

## 阶段 22 - Fuzz 与性质测试引入

目标：对解析、转换、队列、缓存和边界输入多的包引入低维护成本的性质测试。

候选方向：

- 为 `caseconv`、`env`、`priorityqueue`、`lru`、`ttlmap` 评估 fuzz 或 property-style tests。
- 只保留能稳定复现缺陷的 fuzz seed，避免 CI 变慢或 flaky。
- 将 fuzz 发现的边界样本固化为普通回归测试。

完成标准：

- 边界输入多的包有比手写样例更强的输入覆盖。
- fuzz 不进入日常 quick gate，除非成本可控且收益明确。

## 阶段 23 - 性能基线与退化防护

目标：为真正性能敏感的包建立可解释 benchmark，防止无意退化。

任务：

- 识别需要 benchmark 的热点：`lru` eviction、`ttlmap` sweep、`batcher` fan-in、`pool` queueing、`httpx` retry overhead。
- benchmark 只比较稳定操作，不把环境噪声写成结论。
- 对性能改动同时保留正确性测试，避免为数字牺牲语义。

完成标准：

- 每个 benchmark 都有明确问题和解释价值。
- 性能数据能支持 API 或实现选择，而不是装饰性数字。

## 阶段 24 - 内存与 goroutine 泄漏防护

目标：系统性检查会启动 goroutine、timer 或持有 map entry 的包，减少长期运行风险。

任务：

- 覆盖 `batcher`、`debounce`、`fanout`、`pool`、`ttlmap`、`httpx` 的关闭、取消和错误路径。
- 为 timer cleanup、subscriber cleanup、worker shutdown、request body close 建立回归测试。
- 在文档中说明调用者需要承担的 close、stop、cancel 责任。

完成标准：

- 长生命周期对象的释放路径有测试覆盖。
- 用户能从 API 文档判断何时需要主动清理。

## 阶段 25 - 错误模型与 panic 边界统一

目标：让包内错误、panic recovery、错误聚合和 context error 的行为一致。

任务：

- 盘点 `errors`、`taskgroup`、`batcher`、`pool`、`shutdown` 中 panic-to-error 的语义。
- 明确哪些包返回原始错误，哪些包 wrap，哪些包 join。
- 对 context cancellation、deadline exceeded、panic recovery 的优先级补测试。

完成标准：

- 错误行为可以在 README 或 doc comment 中一句话说清楚。
- 不同包之间没有意外相反的错误优先级。

## 阶段 26 - Context 使用规范化

目标：让所有接收 `context.Context` 的 API 都遵循一致、可预期的取消语义。

任务：

- 审查是否存在存储 context、忽略 cancellation、或在后台 goroutine 中错误复用 context 的行为。
- 对 `httpx`、`batcher`、`keyedmutex`、`pool`、`taskgroup`、`shutdown` 的 context path 补边界测试。
- 文档中明确 context 只控制本次调用还是控制对象生命周期。

完成标准：

- context 语义在所有包中一致。
- 取消后的资源释放有测试支持。

## 阶段 27 - 泛型 API 审查

目标：确保泛型包的类型参数、零值、比较器和回调签名清晰，不给用户制造类型推断负担。

任务：

- 审查 `lru`、`ttlmap`、`priorityqueue`、`batcher`、`fanout`、`pool`、`taskgroup` 的泛型签名。
- 检查 zero value 是否可用，不可用时文档明确要求 constructor。
- 补充类型推断友好的 examples。

完成标准：

- 泛型 API 在常见调用中不需要多余类型标注。
- 不可用 zero value 的类型有清晰说明。

## 阶段 28 - 零值与构造器策略统一

目标：决定哪些类型支持 zero value，哪些类型必须通过 constructor，并保持包间一致。

任务：

- 盘点所有 exported structs，记录 zero value 是否安全、是否有用。
- 对必须 constructor 的类型，在 doc comment 和 README 中说明原因。
- 对 builder/options 模式进行一致性审查，避免 nil option 或无效配置行为不一。

完成标准：

- 用户可以从文档判断是否能直接声明变量使用。
- constructor、builder、option 的错误处理模式一致。

## 阶段 29 - 包内私有实现清理

目标：降低长期维护成本，清除重复 helper、过度抽象和未使用路径。

任务：

- 在不改变公开 API 的前提下整理重复的测试 helper、clock helper、stats helper。
- 删除 release 后确认不再需要的兼容残留或死代码。
- 保持重构小步进行，每步都有针对性测试。

完成标准：

- 私有实现更直接，公开行为不变。
- 重构不会和功能新增混在同一个大切片中。

## 阶段 30 - Root 包定位与顶层 doc.go 打磨

目标：让根包和顶层文档准确表达项目定位，不误导用户把它当框架使用。

任务：

- 检查 `doc.go` 是否与 README 中的 package index、stability、dependency boundary 一致。
- 避免在根包暴露跨包 facade，保持按包导入。
- 为 godoc 阅读路径补足必要的简短说明。

完成标准：

- godoc 首页能解释项目是什么、不是什么。
- 根包不成为隐藏依赖聚合点。

## 阶段 31 - 包命名与目录命名审计

目标：在 v1.0 前修正不清晰或容易误解的包名、文件名和术语。

任务：

- 审查 `httpx`、`zapctx`、`slogctx`、`ttlmap`、`defaultmap` 等名称是否与功能边界匹配。
- 对可能改名的包评估迁移成本和收益。
- 若决定保留名称，在 README 中用一句话限定范围。

完成标准：

- v1.0 前没有明显后悔的包名。
- 术语在中英文文档中一致。

## 阶段 32 - README 包表自动校验增强

目标：把 README 包表从人工维护变成强约束，减少新增包时遗漏文档或示例。

任务：

- 扩展现有 smoke tests，校验 package table、实际 package、README、README.zh、example test 的对应关系。
- 检查 optional modules 表与 nested `go.mod` 是否一致。
- 对 examples 表或目录说明增加类似 guard。

完成标准：

- 新增公开包时漏文档、漏中文、漏示例会直接失败。
- 文档表格不再靠人工记忆维护。

## 阶段 33 - Changelog 与 Migration 自动一致性检查

目标：降低 release 文档漂移风险，让 breaking change 和迁移说明互相指向。

任务：

- 检查 `CHANGELOG.md` 与 `CHANGELOG.zh.md` 的 release heading 和主要分区一致。
- 检查 `MIGRATION.md` 与 `MIGRATION.zh.md` 的版本 heading 一致。
- 对 removed/moved packages 建立简单关键词 guard，避免公开文档遗漏。

完成标准：

- 中英文 release/migration 文件结构不漂移。
- breaking/removal 不会只写一边。

## 阶段 34 - 示例模块依赖治理

目标：让 `examples/` 能展示真实组合能力，但不污染 root module，也不成为 CI 盲区。

任务：

- 明确 examples module 的依赖准入标准。
- 为每个 runnable example 确认是否需要 smoke test 或 go test 覆盖。
- 避免 examples 依赖本地环境、外部网络或真实服务。

完成标准：

- examples 可稳定在 split-module gate 中验证。
- examples 不影响 root dependency smoke test。

## 阶段 35 - Benchmark 模块治理

目标：让 benchmark 独立于 root module，同时能服务真实优化决策。

任务：

- 明确 `lru/benchmark` 这类模块的归属、运行方式和 CI 策略。
- 如果新增 benchmark module，同步 CI cache path 和 split gate。
- 不把 benchmark-only 依赖加入 root。

完成标准：

- benchmark 可以按需运行，不拖慢普通 gate。
- benchmark module 依赖边界清晰。

## 阶段 36 - 安全策略与敏感信息处理

目标：系统性处理日志、错误、HTTP、文件路径中的敏感信息风险。

任务：

- 为 `httpx`、`zapctx`、`slogctx`、`rotatingwriter` 审查敏感字段输出。
- 文档中说明库不会替应用做完整脱敏策略，但默认不记录高风险数据。
- 对 query、body、headers、file path 的处理建立 regression tests。

完成标准：

- 默认行为不泄露明显敏感数据。
- 应用自定义脱敏责任边界清楚。

## 阶段 37 - 日志上下文生态收敛

目标：让 `zapctx`、`slogctx` 和 optional integrations 的边界清楚，避免日志包膨胀。

任务：

- 比较 `zapctx` 与 `slogctx` API 形态，尽量保持概念一致。
- 保证 framework middleware 只存在 nested modules。
- 检查 payload logging 的采样、大小限制、跳过逻辑是否足够明确。

完成标准：

- 用户能在 zap/slog 之间迁移概念，而不是学习两套模型。
- 日志集成不会反向污染 core logging context 包。

## 阶段 38 - gRPC 与 HTTP 集成边界审计

目标：让已有 `zapctx` Gin/gRPC 集成稳定，并为未来集成提供模板。

任务：

- 审查 `zapctx/zapctxgin` 与 `zapctx/zapctxgrpc` 的 README、examples、payload tests。
- 确认 middleware/interceptor 不吞掉错误、不破坏 context、不记录过量 payload。
- 将成熟模式抽象为文档模板，而不是 root 代码抽象。

完成标准：

- 新 optional integration 可以复制成熟目录结构和验证方式。
- 现有集成行为有独立模块测试覆盖。

## 阶段 39 - Release Checklist 脚本化

目标：如果人工发布检查继续增长，把重复检查收敛成脚本，而不是依赖记忆。

任务：

- 统计发布前必须手工核对的项目：docs、migration、go sums、CI cache、full gates、tags。
- 只把可自动判断的项目写入脚本，不把主观判断伪装成自动化。
- 在 README 或贡献文档中说明 release check 用法。

完成标准：

- 发布前重复检查可以一条命令完成大部分。
- 脚本失败信息指向具体修复动作。

## 阶段 40 - 多 Go 版本兼容策略

目标：明确项目支持的 Go 版本范围，避免 CI、go.mod 和用户预期不一致。

任务：

- 确认 `go.mod`、CI `GO_VERSION`、README 的版本暗示一致。
- 决定是否需要 oldest-supported Go 版本测试。
- 新增语言特性或标准库 API 前评估最低版本影响。

完成标准：

- 用户能明确知道需要哪个 Go 版本。
- CI 覆盖与支持承诺一致。

## 阶段 41 - 跨平台行为审查

目标：确保文件路径、权限、signal、timer、HTTP 行为在 Linux/macOS/Windows 上尽量可解释。

任务：

- 审查 `rotatingwriter`、`shutdown`、examples 中的平台相关行为。
- 对无法跨平台一致的行为使用 build tags、文档或测试条件明确表达。
- 评估 CI 是否需要补 Windows 或 macOS smoke job。

完成标准：

- 平台差异不会以隐藏 bug 的形式出现。
- 用户能从文档知道哪些行为是平台相关的。

## 阶段 42 - Race 与压力测试成本控制

目标：保留高价值压力测试，同时避免日常开发被慢测试拖垮。

任务：

- 将 stress tests 分为普通可跑、race-only、long-running 三类。
- 给高成本测试增加明确命令、环境变量或 build tag。
- 确认 CI 运行的是稳定集合，不把 flaky 掩盖成偶发失败。

完成标准：

- 压力测试有价值且可维护。
- quick gate 仍然快，full gate 仍然有足够信心。

## 阶段 43 - Issue/PR 模板与贡献入口

目标：如果项目开始接受外部贡献，给贡献者明确边界和验证要求。

任务：

- 增加或更新 issue/PR 模板，要求说明包、API、依赖影响和验证命令。
- 在贡献说明中强调 root dependency boundary 和中文文档同步要求。
- 避免模板过重，只保留能减少维护成本的信息。

完成标准：

- 外部变更更容易落到正确模块和验证路径。
- 维护者不需要反复询问基础信息。

## 阶段 44 - API Deprecation 标注机制

目标：在不立即破坏用户的情况下，为低价值或将被替代的 API 提供清晰退场路径。

任务：

- 使用 Go doc `Deprecated:` 约定标注弃用 API。
- 在 migration docs 中给出替代方案和移除时间窗口。
- 测试弃用 API 在移除前仍保持基本行为。

完成标准：

- 弃用不是突然删除，而是有文档、有替代、有节奏。
- v1.0 后的兼容承诺不会被随意打破。

## 阶段 45 - 包间组合场景验证

目标：验证常见组合不会互相踩边界，而不是只测试单包孤岛。

任务：

- 设计组合测试或 examples：`httpx` + `ratelimit` + `circuitbreaker`，`pool` + `taskgroup` + `shutdown`，`lru` + `ttlmap` 使用模式。
- 组合测试应保持轻量，不变成应用框架。
- 只验证库承诺的交互，不测试虚构业务流程。

完成标准：

- 常见组合路径在发布前被实际编译和运行。
- 包间边界问题能早发现。

## 阶段 46 - 文档语言质量与术语表

目标：让中英文文档长期保持清晰、一致、少口号，避免翻译漂移。

任务：

- 建立常用术语对应：root module、nested module、full gate、quick gate、payload logging、retry budget。
- 清理过度营销或不精确描述，保留工程事实。
- 对中文文档保持自然表达，不机械翻译英文句式。

完成标准：

- 同一概念在不同文件中使用同一术语。
- 文档读起来像维护者写的工程说明。

## 阶段 47 - 用户迁移样例与兼容性夹具

目标：把迁移指南从文字扩展到可编译的小样例，降低破坏性变更风险。

任务：

- 为重要迁移路径保留 before/after snippet 或编译型 example。
- 对 moved optional modules 验证 import path 仍可按 module-aware tooling 正确解析。
- 迁移样例不进入 root dependency graph。

完成标准：

- 用户可以复制迁移样例快速改代码。
- 迁移文档不会描述无法编译的路径。

## 阶段 48 - 内部脚本可维护性审查

目标：让 `scripts/` 自身保持简单、可读、可移植，不成为隐藏复杂系统。

任务：

- 审查 shell scripts 的参数解析、错误信息、路径处理和 help text。
- 避免脚本之间形成难以理解的隐式依赖。
- 对关键脚本行为增加轻量自检，尤其是 CI cache path 和 module discovery。

完成标准：

- 脚本失败时能说明原因和修复方向。
- 本地运行和 CI 运行语义一致。

## 阶段 49 - 项目定位复盘与范围收敛

目标：定期判断 goost 仍然是不是“小型 Go 工具包集合”，防止范围失控。

任务：

- 按包评估：使用场景是否通用、维护成本是否合理、是否已有更合适标准库替代。
- 对功能扩张 backlog 做删减，不把所有想法都转成 API。
- 明确哪些方向永远不做，例如完整 web framework、配置框架、日志框架、ORM 等。

完成标准：

- 项目范围有边界，拒绝项和接纳项一样清楚。
- 新功能进入计划前先通过定位复盘。

## 阶段 50 - 长期路线图复审与下一轮规划

目标：让 50 阶段计划本身保持有用，避免计划变成没人更新的长文档。

任务：

- 每完成一个 release 或重大阶段后，复审阶段顺序、删除过时项、提升新的高价值项。
- 将已完成阶段的结果迁移到公开文档、脚本或测试中，不让 PLAN 成为唯一事实来源。
- 为下一轮 10-20 个实际执行切片排序，确保长期规划能转化为短期工作。

完成标准：

- PLAN 始终能指导下一步实际工作。
- 长期规划与当前代码、文档、CI 状态保持同步。

## Backlog Ideas

- 只有当包存在真实性能声明或风险时才考虑 benchmark，例如 cache eviction、batch fan-in、retry overhead 或 queueing。
- 可考虑为 optional modules 增加 consumer contract tests，确认 import paths 稳定，同时依赖仍留在 root module 之外。
- 如果 v0.4.0 发布前仍需要大量手工跨文件核对，可考虑增加 release checklist script。

## 审查记录（阶段 6-50 执行总结）

### 阶段 6 - 时间/重试/取消语义收敛

逐一追踪 backoff、batcher、debounce、fanout、pool、ttlmap、httpx、ratelimit 的
timer/goroutine 启动点与 cleanup 路径。所有包都在 context 取消或 Stop/Close 时正确
释放 timer 和 goroutine。backoff.Retry 在循环头部检查 ctx.Err()；batcher 在
ctx.Done() 时返回 ctx.Err()；debounce.Stop 取消 pending timer；fanout.Close 关闭
所有 subscriber channel；pool.Close 关闭 work channel 并等待 worker；ttlmap.Close
关闭 stop channel 终止 sweep goroutine；httpx 的 retry timer 用 defer timer.Stop()；
ratelimit.Wait 在 select 的两个分支都调用 t.Stop()。

### 阶段 7 - I/O 与文件系统安全强化

gosec 在 full gate 中保持 0 issues、0 排除运行。httpx.Options.Logger 文档明确说明
"URL query strings and request bodies are not logged"。rotatingwriter 文件权限
测试覆盖 daily log (0600)、size log (0600)、gzip backup (0600)、目录 (0750)。gzip
备份测试验证 .gz 文件正确生成且权限受限。并发写入测试验证 mutex 保护下无数据损坏。

### 阶段 8 - API 形状与 pre-1.0 兼容性审查

逐包审计 exported types：
- 类型命名一致（简单名词），Stats/Snapshot 按包职责正确选用
- Option 模式按复杂度选用：functional options (pool, ttlmap, shutdown)、struct config
  (httpx.Options, circuitbreaker.Config)、builder (lru, batcher, fanout)
- Error sentinel 使用包前缀 ("pool: closed", "batcher: key not found")
- 泛型签名统一：keyed 类型 [K comparable, V any]，单值类型 [T any]
- 无迁移风险。ROADMAP v1.0 audit 对所有包标记 "Keep"。

### 阶段 22 - Fuzz 与性质测试引入

已有 fuzz tests：caseconv（原有）、env（新增 string/int/bool/duration）、
priorityqueue（新增 PushPop/PushMany/Drain）、lru（新增 SetGet/Eviction/Peek/LargeWorkload）、
ttlmap（新增 SetGet/SetDelete/SetMany）。

### 阶段 23 - 性能基线与退化防护

已有 benchmarks：random.String（原有）、lru Set/Get/UnsafeSet/UnsafeGet（lru/benchmark 模块）、
pool Schedule/ScheduleWithQueue（新增）、batcher Load（新增）、ratelimit BucketAllow/BucketWait（新增）。

### 阶段 24 - 内存与 goroutine 泄漏防护

所有启动 goroutine/timer 的包（batcher、debounce、fanout、pool、ttlmap）都有显式
Close/Stop 方法和对应测试。batcher 的 Test_StressCancelAndTimeout 验证了大量并发
context 取消后的行为。debounce 的 Test_StopCancelsPending 验证了 Stop 取消 pending
timer。fanout 的 Test_BroadcasterCloseClosesAllSubs 验证了 Close 关闭所有 subscriber。
pool 的 Test_CloseDrainsAcceptedQueuedTasks 验证了 Close 等待所有已接受任务完成。
ttlmap 的 Test_CloseStopsSweepButMapRemainsUsable 验证了 Close 停止 sweep 但 map
仍可用。

### 阶段 25 - 错误模型与 panic 边界统一

panic-to-error 路径：taskgroup 用 fmt.Errorf("taskgroup: panic: %v", r)；
batcher 用 fmt.Errorf("batcher: panic in loadFn: %v\n%s", r, debug.Stack())；
pool 通过 WithPanicHandler 暴露 panic 值；shutdown 通过 recover() 吞掉 panic。
这些边界已在 API_CONVENTIONS.md "Errors and Panics" 节记录。

### 阶段 26 - Context 使用规范化

所有接收 context.Context 的 API 遵循 API_CONVENTIONS.md "Context, Cancellation,
and Lifecycle" 节的约定：context 控制当前操作（非对象生命周期），取消后返回
ctx.Err()，Close/Stop 安全幂等。

### 阶段 27 - 泛型 API 审查

所有泛型包（lru, ttlmap, priorityqueue, batcher, fanout, defaultmap）使用
[K comparable, V any] 或 [T any] 约束。不可用 zero value 的类型在 doc comment 中
明确标注（"The zero value is not usable"）。类型推断在常见调用中不需要多余标注。

### 阶段 28 - 零值与构造器策略统一

盘点结果：backoff.Backoff（零值可用）、clock.Clock（零值不可用，需 Real/NewMock）、
lru.Cache（需 New+Builder）、pool.Pool（需 NewPool）、ratelimit.Bucket/Leaky
（零值不可用，需 NewBucket/NewLeaky）、keyedmutex.Mutex（零值不可用，需 New）、
debounce.Debouncer（需 New）、fanout.Broadcaster（需 New+Builder）。
所有不可用零值的类型在 doc comment 中有说明。

### 阶段 29 - 包内私有实现清理

未发现重复 helper、死代码或陈旧兼容残留。rg 扫描未命中 deprecated/unused/TODO:remove
标记。errors 包的 "compatible" 注释是说明标准库兼容性（非残留代码）。shutdown 的
"backwards compatibility" 注释是 API 文档（非内部兼容层）。

### 阶段 13-18 - 功能扩张评估结论

所有候选方向保持在 backlog，理由：
- 阶段 14（cache/data structure）：标准库 map + lru/ttlmap/defaultmap 组合已覆盖
  大部分场景。通用 cache interface 会增加维护负担但收益不明确。
- 阶段 15（concurrency coordination）：keyedmutex + pool + taskgroup 已覆盖常见
  协调模式。semaphore 可用 buffered channel 模拟，不值得独立包。
- 阶段 16（HTTP expansion）：httpx 当前 retry/ratelimit/circuitbreaker 组合已
  足够。metrics hooks 等待真实用户需求再设计。
- 阶段 17（config/startup）：env + shutdown 当前 API 表面足够。required 字段和
  duration/slice 类型可后续按需添加。
- 阶段 18（optional integrations）：zapctx/gRPC/Gin 集成已稳定。slog 集成等
  用户需求明确后再评估。

### 阶段 19-20, 30-50 - 策略与审计

- 阶段 19：v1.0 readiness list 在 ROADMAP.md，keep/change/remove 决策齐全
- 阶段 20：maintenance review 在 ROADMAP.md "Maintenance Review" 节
- 阶段 21：test matrix 在 TESTING.md
- 阶段 30：root doc.go 与 README 一致
- 阶段 31：包命名在 ROADMAP v1.0 Package Audit 表中全部 "Keep"
- 阶段 32：smoke tests 校验 package table/README/example 对应关系
- 阶段 33：smoke tests 校验 CHANGELOG 中英文一致性
- 阶段 34：split-module gate 覆盖 examples
- 阶段 35：lru/benchmark 独立模块
- 阶段 36：gosec 0 issues，httpx 不记录敏感数据
- 阶段 37：zapctx/slogctx API 对齐（zapctx.S 为 zap 特有，无 slog 等价物）
- 阶段 38：zapctxgin/zapctxgrpc 有 README、example test、payload test
- 阶段 39：check-release.sh 存在
- 阶段 40：go.mod 1.25.10 与 CI GO_VERSION 一致
- 阶段 41：shutdown 有 //go:build !windows 测试
- 阶段 42：stress tests 在独立文件（*_stress_test.go, stress_test.go）
- 阶段 43：.github/ISSUE_TEMPLATE/ 和 pull_request_template.md 存在
- 阶段 44：无 deprecated API（pre-1.0 阶段）
- 阶段 45：examples/ 有 cache/concurrent/resilientclient 组合示例
- 阶段 46：术语在 API_CONVENTIONS.md 中统一
- 阶段 47：testdata/migration/ 有 zapctx 迁移 fixture
- 阶段 48：check-scripts.sh 验证脚本质量
- 阶段 49：PROJECT_POLICY.md 记录范围边界
- 阶段 50：ROADMAP.md 当前且与代码同步
