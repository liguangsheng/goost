# 更新日志

本文件记录该项目的显著变更。格式遵循
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/)。

## [Unreleased]

该版本计划作为 v0.4.0 发布，因为它移除了低价值公开包，并把可选/demo
module 从根 module 依赖图中拆出。

### Added

- 新增中文 Markdown 文档：根部文档和每个包 README 都有对应的 `*.zh.md`。
- 新增 consumer dependency smoke test：自动发现核心包集合，并防止核心包导入
  可选的 Gin、gRPC 或 OpenTelemetry 集成依赖。
- 新增编译型 `httpx` 示例，同时覆盖 retry callback 和请求摘要日志。
- `httpx.Options.Logger` 会在重试结束后记录一条脱敏请求摘要，包含 status、
  attempts、duration 和 error，不包含 query string 或 body。
- `httpx.RetryPolicy.OnRetry` 会在下一次请求发送前报告可重试 attempt 的
  status、error、attempt count、next delay 和脱敏请求元数据。

### Changed

- `examples/`、`lru/benchmark`、`zapctx/zapctxgin` 和 `zapctx/zapctxgrpc`
  现在拥有各自的 `go.mod`，避免 demo、benchmark、Gin 和 gRPC 依赖进入根库
  module。
- `scripts/check-root.sh` 和 `scripts/check-split-modules.sh` 现在分别是
  root 与 nested module 在本地和 CI 中共享的检查入口，支持 `--quick`、
  `--full` 和指定 module 检查。Full gate 覆盖 tidy、vet、tests、lint/静态分析、
  漏洞检查和安全扫描。
- CI 现在显式运行 root 与 nested module 的 full gate，通过
  `scripts/install-ci-tools.sh` 安装工具，在 workflow environment 中集中声明
  Go/工具版本，并使用 root 与所有 nested module 的 `go.sum` 作为 Go module
  cache key。
- Root 检查现在会验证 CI `cache-dependency-path` 是否与仓库中的 `go.sum`
  文件保持一致。
- 示例不再触发安全检查：HTTP server 配置了 `ReadHeaderTimeout`，concurrent
  retry demo 改用确定性的临时失败，不再使用 `math/rand`。
- CI 现在使用 Node 24-native 的 `actions/checkout@v6` 和
  `actions/setup-go@v6`。
- CI 现在使用 Node 24-native 的 `codecov/codecov-action@v6`。

### Removed

- 移除低价值公开包：`bytesconv`、`itertools`、`redact`、
  `slogctx/slogctxotel` 和 `zapctx/zapctxotel`。
- `random` 不再依赖 `bytesconv`；随机字符串生成改用普通
  `string([]byte)` 转换。

### Fixed

- `httpx` 现在会在每次 retry attempt 前应用 `Options.Limiter`，而不是只在
  第一次请求前执行一次。
- `httpx` 现在会通过 `Request.GetBody` 为 retry attempt 回放请求 body，
  而不是依赖已经被消费过的 body。

## [v0.3.0] - 2026-05-21

v0.3.0 完成了 v0.2.0 后开始的日志集成拆分，为新的公开入口补充了可编译示例，
并加强了并发覆盖。API 仍处于 pre-1.0。

### Added

- 为 `batcher`、`clock.Mock`、`debounce`、`errors.Recover`、`fanout`、
  `keyedmutex` 和 `priorityqueue` 增加 example tests，确保公开示例会被
  `go test` 编译。
- 为 `batcher`、`fanout`、`keyedmutex`、`pool` 和 `ttlmap` 增加面向 race 的
  stress tests。
- 新增 `zapctx/zapctxgin`、`zapctx/zapctxgrpc`、`zapctx/zapctxotel` 和
  `slogctx/slogctxotel` 子包，用于可选框架和 tracing 集成。
- 新增 v0.3.0 迁移说明，覆盖移动后的日志集成 API。

### Changed

- 将 Gin、gRPC 和 OpenTelemetry 集成从核心 `zapctx` 中移出，并将
  OpenTelemetry hook 从核心 `slogctx` 中移出；只导入核心日志 context 包时，
  不再编译这些可选集成。
- `batcher.New`/`Build` 现在会立即拒绝 nil load function。
- `debounce.WithClock(nil)` 和 `ttlmap.New(..., nil)` 现在会忽略 nil 配置值，
  而不是保存它们。
- `fanout.Builder.Build` 现在会在 builder 意外留下无效 buffer size 时恢复默认
  buffer。
- CI 提前切换到 Node.js 24，以应对 GitHub Node 20 runner deprecation。

## [v0.2.0] - 2026-05-21

该版本把 goost 从一个小型工具集合扩展成更偏生产使用的并发、调度、日志和通用工具包。
同时移除了低价值的 `singleflight` 包装。API 仍处于 pre-1.0。

### Added

- **`batcher`**：新包。DataLoader 风格，将并发的按 key `Load(ctx, key)`
  调用合并成一次批量 `loadFn` 调用；支持 `MaxBatch` / `MaxWait`、
  panic-to-error、缺失 key 的 `ErrNotFound` 和用于调优的 `Stats()`。
- **`keyedmutex`**：新包。按 key 互斥，提供 `Lock` / `TryLock` /
  `LockContext` / `WithLock` / `Len`。slot 惰性分配，并在没有持有者或等待者时释放。
- **`clock`**：`Clock` 接口新增 `AfterFunc(d, fn) Timer` 和
  `NewTicker(d) Ticker`，两者都可 mock。Mock 可确定性驱动 ticker。
- **`fanout`**：新包。进程内广播器，将每个 `Publish(v)` 发送给所有当前
  `Sub`。慢订阅者采用 drop-on-full 背压，提供每订阅者 `Drops()` 和聚合 `Stats()`。
- **`errors`**：`Recover(*error)` 可把 deferred `recover()` 转成 `*PanicError`。
  `PanicError` 保留 panic value 和 `debug.Stack()` 快照；`%+v` 会打印两者。
- **`rotatingwriter`**：`DailyRotater` 和 `SizeRotater` 都支持
  `WithMaxAge(d)`。轮转时如果备份超过数量或年龄限制，会被删除。
- **`priorityqueue`**：新包。基于 `container/heap` 的泛型最小/最大堆，
  使用 `New(less)`，不需要为每个类型实现五个方法。支持 `Push` / `Pop` /
  `Peek` / `Len` / `Clear` / `Drain`。
- **`debounce`**：新包。`Debouncer[T any]` 把一串 `Trigger(v)` 合并成安静窗口后
  从 `C()` 发送的一次事件。慢消费者采用 latest-wins；可注入 `clock.Clock`。
- **`lru`**：`Cache` 和 `ShardedCache` 都新增 `Keys()` 和 `Range(fn)`，
  跳过过期条目；迭代按每个 shard 的 MRU-first 顺序。
- **`defaultmap`**：`GetOrInit` 返回 `(V, loaded bool)`；新增
  `LoadOrStore`，实现类似 `sync.Map` 的插入且不会调用构造函数。
- **`zapctx`**：`PayloadGinMiddleware` 记录 request/response body、status 和
  latency，支持 `WithMaxBody`、`WithSampling`、`WithSkipper`；gRPC 也有对应
  `PayloadUnaryServerInterceptor`。
- **`pool`**：`Stats()` 通过轻量 atomic 暴露 workers、in-flight、queued、
  completed 和 panic counts。
- **`ttlmap`**：`WithOnExpire` hook 会在条目因访问或后台扫描被驱逐时触发。
- **`random`**：`SecureString` 使用 `crypto/rand`，适合 token/salt。
- **`backoff`**：`Backoff.Rand` 允许测试注入确定性 jitter source。
- **`caseconv`**：新增一步式 `ToUpperCamel` / `ToLowerSnake` /
  `ToLowerKebab` 等，自动检测输入风格。
- **`slogctx`**：新包。`zapctx` 的 `log/slog` 对应实现，包含 `OtelTraceInject`。
- **`errors`**：新包。兼容 `errors.Is`/`As`/`%w` 的轻量 stack-trace wrapping。
- **`taskgroup`**：新包。`errgroup` + 并发限制 + panic 恢复。
- **`ratelimit`**：新包。token bucket 和 leaky bucket，支持 `Allow` /
  `Wait(ctx)` 和可注入时钟。
- **`circuitbreaker`**：新包。closed / open / half-open 状态机，支持可配置阈值、
  冷却时间和 `OnStateChange`。
- **`clock`**：新包。`Clock` 接口包含 `Real` 和 `Mock`（`Advance` / `Set`）；
  `Mock.Now` 可接入已有模块的 `SetClock`。
- **`httpx`**：新包。组合 `backoff` retry、`ratelimit`、`circuitbreaker` 的
  `*http.Client` 工厂；会缓冲 request body，使其可在 retry 中复用。
- **`env`**：新包。基于 struct tag 的环境变量配置加载器，支持 `default=`、
  `required` 和自定义 map。
- **`redact`**：新包。日志友好的字符串脱敏（`Mask`、`Email`、`Phone`、
  `Token`），以及 `ZapString` / `SlogString` 字段辅助函数。
- **`lru`**：`Resize(n)` 支持运行时调整容量，缩小时驱逐 LRU 条目。
- **`pool`**：`ScheduleN([]task)` 可调度一批任务。
- **`taskgroup`**：`Results[T]` 可在第一个错误之外收集并发任务的成功返回值。
- **`errors`**：`Join(errs...)` 在 `stderrors.Join` 外附加 stack；
  `JoinFormatPlusV` 将每个 joined error 单独打印一行。
- 根 `doc.go` 总结 module 包列表。
- 新增 `examples/` 目录，包含三个可运行程序（`httpserver`、`concurrent`、
  `cache`），演示多个包的组合。
- CI 现在运行 Go 1.25.10，并包含 `staticcheck`、`gosec`、`golangci-lint`
  和 `govulncheck` jobs。

### Removed

- **`singleflight`**：移除该包。它只是 `golang.org/x/sync/singleflight` 的薄泛型包装，
  没有足够附加价值；请直接依赖 `x/sync/singleflight`。`examples/cache`
  现在直接导入它。对于“不同 key 的请求合并”（singleflight 不处理），请看新的
  [`batcher`](./batcher) 包。

## [v0.1.0] - 2026-05-20

第一个带 tag 的版本。两个阶段的内部清理合并成一个基线版本。API 仍处于 pre-1.0；
可能发生破坏性变更。

### Added

- **`lru`**：泛型 `Cache[K, V]`，支持 `Peek`、evict hook 和单条目纳秒级过期。
  新的 `ShardedCache` 通过 `Builder.Shards` 和 `BuildSharded` 降低锁竞争；
  提供 `StringHash` 辅助函数。
- **`itertools`**：新增 `Intersection`、`Contains`、`Chunk`。
  （`Reject` 重命名为 `Intersection`。）
- **`defaultmap`**：新增 `Has`、`Len`、`Range`。
- **`pool`**：新增 `Close`、`WithPanicHandler` 选项和完整参数校验。不再依赖 zap。
- **`shutdown`**：新增 `Manager`、`Wait(ctx)`、`SetLogger`、per-hook
  `WithTimeout`/`WithName` 选项；hook panic 会被恢复。
- **`zapctx`**：新增 `OtelTraceInject`（OpenTelemetry）；`BetterDefault`
  现在返回 error；导出 `ZapContext`。
- **`rotatingwriter`**（由 `rotating_writer` 重命名）：新增 `SizeRotater`，
  支持可选 gzip 备份；`Write` 并发安全。
- **`bytesconv`**：现代化为 `unsafe.String` / `unsafe.SliceData`。
- **`backoff`**：新包。指数退避、jitter、`Retry(ctx, ...)`、`Permanent(err)`。
- **`singleflight`**：新包。`golang.org/x/sync/singleflight` 的泛型包装。
- **`ttlmap`**：新包。带单条目 TTL 和可选后台扫描的并发 map。
- 为 lru、itertools、caseconv、defaultmap 增加 `ExampleXxx`。
- 为 caseconv split/join round-tripping 增加 fuzz tests。
- 新增 `.golangci.yml`（v2 schema）；GitHub Actions CI matrix（Go 1.24、1.25），
  并上传 codecov。
- `lru/benchmark` 现在与 `golang-lru/v2` 和 `ristretto` 对比。

### Changed

- Go directive 提升到 `1.23`（toolchain `1.25`）。
- `caseconv`：替换已废弃的 `strings.Title`；常量重命名为 Go 惯用命名；
  `AcronymMap` 改为通过 `RegisterAcronym`/`UnregisterAcronym`/`IsAcronym` 隐藏。
- `random`：使用 `math/rand/v2`；默认并发安全。
- `rotatingwriter`：包名不再包含下划线。

### Removed

- `lru/list.go`（自实现双向链表）被 `container/list` 替代。
- `lru.LRU` type alias 和私有 `lru` 类型由 `Cache[K, V]` 取代。
- `zapctx.OpenTraceInject` 和 `go.opencensus.io` 依赖。
- `labstack/gommon` 间接依赖。

### Fixed

- `lru.Get`/`Peek` 现在会在访问时驱逐过期条目。
- `lru.Clear` / `lru.Size` 会获取锁。
- `random.Sequence.Next` 在并发首次调用时是安全的。
- `random.String(0, ...)` 不再 underflow。
- `rotatingwriter.DoRollover` 在打开新文件失败时保留旧文件。
- `zapctx.OtelTraceInject`（以及之前的 `OpenTraceInject`）在 context 没有关联
  logger 时不再 panic。
- 修复此前无法编译的测试代码（lru tests 中的 `string(int)`）。

[v0.2.0]: https://github.com/liguangsheng/goost/releases/tag/v0.2.0
[v0.1.0]: https://github.com/liguangsheng/goost/releases/tag/v0.1.0
