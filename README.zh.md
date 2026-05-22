# goost

作者在项目中反复会用到的一组小型 Go 工具。

## 安装

```sh
go get github.com/liguangsheng/goost
```

## 包

| 包 | 用途 |
| --- | --- |
| [`backoff`](./backoff) | 带 jitter 的指数退避，以及感知 `context` 的 `Retry`。 |
| [`batcher`](./batcher) | DataLoader 风格：把并发的按 key 加载合并成一次批量调用。 |
| [`caseconv`](./caseconv) | camel/snake/kebab/pascal case 拆分与拼接辅助函数。 |
| [`circuitbreaker`](./circuitbreaker) | 三态熔断器（closed/open/half-open），保护下游调用。 |
| [`clock`](./clock) | `Clock` 抽象，包含 `Real` 和 `Mock`，用于确定性时间测试。 |
| [`debounce`](./debounce) | 把一串 `Trigger(v)` 合并成安静窗口后的单次发送；latest-wins。 |
| [`defaultmap`](./defaultmap) | 并发 map，缺失 key 时惰性构造值。 |
| [`env`](./env) | 基于 struct tag 的环境变量配置加载器。 |
| [`errors`](./errors) | 带 stack trace 的 `errors`；`Join`；`Recover`；兼容 `Is`/`As`/`%w`。 |
| [`fanout`](./fanout) | 进程内广播器：一个发布者，多个订阅者；慢订阅者丢弃。 |
| [`httpx`](./httpx) | 带 retry、ratelimit、circuit breaker 和请求日志的 `*http.Client`。 |
| [`keyedmutex`](./keyedmutex) | 按 key 互斥：不同 key 并行，相同 key 串行；空闲 slot 自动 GC。 |
| [`lru`](./lru) | 泛型 LRU 缓存，支持可选单条目过期；分片变体；`Keys`/`Range`/`Resize`。 |
| [`pool`](./pool) | 有界 goroutine 池，支持可选队列、panic handler 和 `Stats`。 |
| [`priorityqueue`](./priorityqueue) | 基于 `container/heap` 的泛型最小/最大堆，用 comparator 替代五个方法。 |
| [`random`](./random) | 并发安全随机字符串生成器；`SecureString` 使用 `crypto/rand`。 |
| [`ratelimit`](./ratelimit) | token bucket 和 leaky bucket 限流器。 |
| [`rotatingwriter`](./rotatingwriter) | 会轮转底层文件的 `io.Writer`（按天或按大小，可选 gzip）。 |
| [`shutdown`](./shutdown) | 信号驱动的优雅关闭协调器（支持 per-hook timeout）。 |
| [`slogctx`](./slogctx) | 通过 `context.Context` 携带 `*slog.Logger` 和 attrs。 |
| [`taskgroup`](./taskgroup) | `errgroup` + 并发限制 + panic 恢复。 |
| [`ttlmap`](./ttlmap) | 带单条目过期时间和后台扫描的并发 map。 |
| [`zapctx`](./zapctx) | 通过 `context.Context` 携带 `*zap.Logger` 和结构化字段。 |

可运行的端到端程序位于 [`examples/`](./examples)。

所有包彼此独立；按需依赖即可。

## 可选集成 Modules

| Module | 用途 |
| --- | --- |
| [`zapctx/zapctxgin`](./zapctx/zapctxgin) | `zapctx` 的 Gin 中间件和 HTTP payload 日志。 |
| [`zapctx/zapctxgrpc`](./zapctx/zapctxgrpc) | `zapctx` 的 gRPC interceptor 和 payload 日志。 |

## 稳定性

该 module 仍处于 pre-1.0。API 可能在 minor version 之间变化。
发布历史见 [CHANGELOG.zh.md](./CHANGELOG.zh.md)。
minor version 迁移说明见 [MIGRATION.zh.md](./MIGRATION.zh.md)。

## 开发检查

日常 root module 改动先运行 `./scripts/check-root.sh --quick`。Nested modules
是独立 Go modules，root 检查不会遍历它们。
Root gate 也会验证每个列在 README 中的公开包都有已编译的 `Example` 测试；
`examples/` 中的可运行程序则由 split-module gate 覆盖。

CI 通过 `./scripts/install-ci-tools.sh` 安装所需分析工具，工具版本由 workflow
environment 统一控制。

发布前运行：

```sh
./scripts/check-root.sh --full
./scripts/check-split-modules.sh --full
```

日常 nested module 改动则针对改过的 module 运行 split-module gate：

```sh
./scripts/check-split-modules.sh --quick --module zapctx/zapctxgin
```

## License

MIT，见 [LICENSE](./LICENSE)。
