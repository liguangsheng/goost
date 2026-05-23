# 路线图

本路线图把长期规划转成公开检查点。它不是日期承诺。

## v1.0 readiness

v1.0 之前，项目应具备：

- 每个公开包都有最终的 keep/change/remove 决策；
- 对 `httpx`、`zapctx`、`slogctx`、`ttlmap`、`defaultmap` 等短名称或易混名称作出命名决策；
- 每个 v1.0 前仍计划发生的 breaking change 都有迁移说明；
- root、examples、benchmarks 和 optional integrations 的依赖边界稳定；
- 每个公开包和 nested module 都有可编译示例和中英文 README；
- root 与 split-module full gates 能通过 `./scripts/check-release.sh` 跑通；
- 兼容性、弃用、安全、测试和 API 约定都有公开文档。

## v1.0 package audit

本表跟踪公开包当前的 v1.0 方向。`Keep` 表示目前符合项目范围。`Review` 只保留给 v1.0 前仍需要明确包名或 API 形状决策的包。

| Package | Direction | v1.0 note |
| --- | --- | --- |
| `backoff` | Keep | 保持 retry/backoff primitives 小而清晰，并继续感知 context。 |
| `batcher` | Keep | 保持 DataLoader 风格的按 key load coalescing。 |
| `caseconv` | Keep | 保持字符串 case conversion helpers，不扩展到 locale-heavy 范围。 |
| `circuitbreaker` | Keep | 保持为小型下游保护 primitive。 |
| `clock` | Keep | 保持为 deterministic timer 测试抽象。 |
| `debounce` | Keep | 保持 latest-wins quiet-window 语义明确。 |
| `defaultmap` | Keep | 保留该名称；README 和 examples 已明确 lazy construction 语义。 |
| `env` | Keep | 保持小型 struct-tag environment loader，不变成配置框架。 |
| `errors` | Keep | 保持兼容标准 `errors.Is` / `errors.As`。 |
| `fanout` | Keep | 保持进程内 broadcaster 语义和 drop-on-slow 行为明确。 |
| `httpx` | Keep | 保留短名称，并把范围限定为 outbound HTTP client assembly。 |
| `keyedmutex` | Keep | 保持 per-key mutual exclusion 和 idle slot cleanup。 |
| `lru` | Keep | 保持 generic LRU/cache primitives，benchmark 独立维护。 |
| `pool` | Keep | 保持 bounded goroutine pool 行为可观测且 lifecycle-safe。 |
| `priorityqueue` | Keep | 保持 comparator-based heap wrapper 小而清晰。 |
| `random` | Keep | 保持 random string helpers 与 secure generation 边界清楚。 |
| `ratelimit` | Keep | 保持 token/leaky bucket 在 context cancellation 下行为可预测。 |
| `rotatingwriter` | Keep | 保持 file permissions、retention 和 gzip 行为明确。 |
| `shutdown` | Keep | 保持 signal-driven graceful shutdown coordination 范围克制且可移植。 |
| `slogctx` | Keep | 保持为 `zapctx` 的 `log/slog` 对应包，并已记录共享概念。 |
| `taskgroup` | Keep | 保持 errgroup-style concurrency、limit 和 panic recovery。 |
| `ttlmap` | Keep | 保留该名称；close、lazy expiration、sweep 和 purge 语义已有文档和测试。 |
| `zapctx` | Keep | 保留命名，并继续把 framework integrations 留在 nested modules。 |

## 维护复审

每个 release boundary 都要复审：

- package names、package scope 和 keep/change/remove 决策；
- 依赖和 Go toolchain 是否漂移；
- CI action versions 和 cache paths 是否仍然正确；
- 文档链接、examples 和 migration snippets 是否 stale 或损坏；
- 是否存在应弃用或移除的低价值 API；
- backlog ideas 是否仍符合项目范围。

持久决策应进入公开文档、脚本或测试，而不是只留在私有计划里。

复审后，应排序下一轮 10 到 20 个具体执行切片。每个切片都要有负责的改动面、预期制品和验证命令。宽泛想法必须先通过项目政策中的新增准入标准，才能进入 roadmap。

## Next Execution Slices

11-20 切片已完成。下表是下一轮的当前执行队列。

| Slice | Surface | Artifact | Validation |
| --- | --- | --- | --- |
| 21 | `lru` observability wording | 统一 `lru` 和 `ShardedCache` 的 Snapshot 表述 | `./scripts/check-root.sh --quick` |
| 22 | `circuitbreaker` observability wording | 统连续计数器和时间戳的 Snapshot 表述 | `./scripts/check-root.sh --quick` |
| 23 | `ratelimit` stress tests | 为 Bucket 和 Leaky 添加并发 Wait stress tests | `go test -race ./ratelimit` |
| 24 | `debounce` concurrent tests | 添加并发 Trigger + Stop stress test | `go test -race ./debounce` |
| 25 | `pool` cancellation tests | 添加 pre-cancelled context 下的 Schedule/ScheduleTimeout 测试 | `go test -race ./pool` |
| 26 | `fanout` concurrent edge tests | 添加并发 publish 期间 subscribe/close 测试 | `go test -race ./fanout` |
| 27 | package doc parity | 为 `caseconv` 和 `random` 添加 doc.go | `go vet ./caseconv ./random` |
| 28 | stress gate refresh | 21-27 切片后重新运行并记录 `./scripts/check-stress.sh` | `./scripts/check-stress.sh` |
| 29 | release gate refresh | 21-28 切片后重新运行并记录 `./scripts/check-release.sh` | `./scripts/check-release.sh` |
