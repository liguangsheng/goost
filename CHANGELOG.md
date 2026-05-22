# Changelog

All notable changes to this project will be documented in this file.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

This is planned as v0.4.0 because it removes low-value public packages and
splits optional/demo modules out of the root module dependency graph.

### Added

- Chinese-language Markdown docs now mirror the root docs and every package
  README as `*.zh.md`.
- A consumer dependency smoke test now discovers the core package set and
  guards it from importing optional Gin, gRPC, or OpenTelemetry integration
  dependencies.
- A compiled `httpx` example now covers retry callbacks and request summary
  logging together.
- `httpx.Options.Logger` logs one sanitized request summary after retries
  finish, including status, attempts, duration, and error without query
  strings or bodies.
- `httpx.RetryPolicy.OnRetry` reports retryable attempts with status, error,
  attempt count, next delay, and sanitized request metadata before the next
  request is sent.
- `httpx.RetryPolicy.OnGiveUp` reports the final retryable response or
  transport error when the retry budget is exhausted.
- `circuitbreaker.Breaker.Snapshot` exposes a read-only state and cooldown
  view for metrics and logs.
- `ratelimit.Bucket.Snapshot` and `ratelimit.Leaky.Snapshot` expose read-only
  limiter state for metrics and logs.
- `ttlmap.TTLMap.PurgeExpired` removes expired entries on demand, reports the
  removal count, and fires expiration hooks.
- `lru.Cache.Snapshot` and `lru.ShardedCache.Snapshot` expose read-only size,
  capacity, and shard-count views for metrics and logs.
- `batcher.Stats` now includes open-window size, in-flight load count, and
  configured batch limits for tuning and observability.
- `fanout.Stats` now includes buffer size, queued message count, and closed
  state for runtime observability.
- `pool.Stats` now includes worker capacity, queue capacity, and closed state
  for runtime observability.

### Changed

- `examples/`, `lru/benchmark`, `zapctx/zapctxgin`, and `zapctx/zapctxgrpc`
  now have their own `go.mod` files, keeping demo, benchmark, Gin, and gRPC
  dependencies out of the root library module.
- `scripts/check-root.sh` and `scripts/check-split-modules.sh` now provide the
  shared local/CI gates for the root and nested modules, with `--quick`,
  `--full`, and targeted module checks. Full gates cover tidy, vet, tests,
  lint/static analysis, vulnerability checks, and security scanning.
- Full local/CI gates now run `gosec ./...` without rule exclusions.
- `rotatingwriter` now creates log directories/files with more restrictive
  default permissions and documents intentional caller-selected file paths.
- CI now explicitly runs the full root and nested-module gates, installs tools
  through `scripts/install-ci-tools.sh`, declares Go/tool versions once in the
  workflow environment, and keys Go module caching off every root and nested
  module `go.sum` file.
- Root checks now verify that CI `cache-dependency-path` entries stay aligned
  with repository `go.sum` files.
- Examples no longer trip security checks: the HTTP server configures
  `ReadHeaderTimeout`, and the concurrent retry demo uses deterministic
  transient failures instead of `math/rand`.
- CI now uses Node 24-native `actions/checkout@v6` and `actions/setup-go@v6`.
- CI now uses Node 24-native `codecov/codecov-action@v6`.

### Removed

- Removed low-value public packages: `bytesconv`, `itertools`, `redact`,
  `slogctx/slogctxotel`, and `zapctx/zapctxotel`.
- `random` no longer depends on `bytesconv`; random string generation now uses
  the ordinary `string([]byte)` conversion.

### Fixed

- `httpx` retry handling now isolates backoff state per request, stops retry
  delay timers promptly on context cancellation, and leaves the final response
  body open for the caller while still closing intermediate retry responses.
- `httpx` now applies `Options.Limiter` before each retry attempt instead of
  only once before the first request.
- `httpx` now replays request bodies through `Request.GetBody` on each retry
  attempt instead of relying on a previously consumed body.
- `batcher.Stats().MaxBatchSize` now remains monotonic under overlapping batch
  executions.
- `rotatingwriter.DailyRotater` now treats `maxBackup` as the number of
  historical backup files to keep, matching size-based rotation.
- `taskgroup.Group.Cause()` now returns the first task error directly, and
  `taskgroup.Results[T]` now exposes the same `Cause()` helper.
- `random` no longer triggers integer-conversion warnings in security scans.
- `batcher` now flushes immediately when the first key in a new window reaches
  `MaxBatch`, including the `MaxBatch(1)` case.

## [v0.3.0] — 2026-05-21

v0.3.0 finishes the logging-integration split started after v0.2.0, adds
compiled examples for the new public entry points, and tightens concurrency
coverage. APIs remain pre-1.0.

### Added

- Example tests for `batcher`, `clock.Mock`, `debounce`, `errors.Recover`,
  `fanout`, `keyedmutex`, and `priorityqueue`, so public examples are
  compiled by `go test`.
- Race-oriented stress tests for `batcher`, `fanout`, `keyedmutex`, `pool`,
  and `ttlmap`.
- `zapctx/zapctxgin`, `zapctx/zapctxgrpc`, `zapctx/zapctxotel`, and
  `slogctx/slogctxotel` subpackages for optional framework and tracing
  integrations.
- v0.3.0 migration notes covering the moved logging integration APIs.

### Changed

- Moved Gin, gRPC, and OpenTelemetry integrations out of core `zapctx`, and
  moved the OpenTelemetry hook out of core `slogctx`, so importing the core
  logging context packages no longer compiles those optional integrations.
- `batcher.New`/`Build` now reject a nil load function immediately.
- `debounce.WithClock(nil)` and `ttlmap.New(..., nil)` now ignore nil
  configuration values instead of storing them.
- `fanout.Builder.Build` now restores the default buffer if a builder is
  somehow left with an invalid buffer size.
- CI opts into Node.js 24 for JavaScript actions ahead of GitHub's Node 20
  runner deprecation.

## [v0.2.0] — 2026-05-21

This release expands goost from a small grab bag into a broader set of
production-leaning concurrency, scheduling, logging, and utility packages.
It also removes the low-value `singleflight` wrapper. APIs remain pre-1.0.

### Added

- **`batcher`** — new package: DataLoader-style coalescing of concurrent
  per-key `Load(ctx, key)` calls into a single batch `loadFn` invocation,
  with `MaxBatch` / `MaxWait`, panic-to-error, `ErrNotFound` for keys
  missing from the batch result, and `Stats()` for tuning.
- **`keyedmutex`** — new package: per-key mutex with `Lock` / `TryLock` /
  `LockContext` / `WithLock` / `Len`. Slots are allocated lazily and
  freed once no goroutine holds or waits on the key, so a churn of
  many one-shot keys does not grow the internal map.
- **`clock`** — `Clock` interface gained `AfterFunc(d, fn) Timer` and
  `NewTicker(d) Ticker`; both are mockable. Mock now drives tickers
  deterministically (one tick per period boundary crossed during
  `Advance`/`Set`; missed ticks drop, matching `time.Ticker`).
- **`fanout`** — new package: in-process broadcaster delivering each
  `Publish(v)` to every current `Sub`. Drop-on-full backpressure (slow
  subscribers never block publishers), per-sub `Drops()` counter,
  aggregate `Stats()`.
- **`errors`** — `Recover(*error)` turns a deferred `recover()` into
  a `*PanicError`. PanicError preserves the panic value and a
  `debug.Stack()` snapshot; `%+v` prints both. If the named-return
  error is already non-nil, the PanicError is joined via
  `errors.Join`.
- **`rotatingwriter`** — both `DailyRotater` and `SizeRotater` now
  support `WithMaxAge(d)`. A backup is deleted at rollover if either
  count (`maxBackup`) or age limit would be exceeded. Daily uses the
  date encoded in the filename; size uses mtime.
- **`priorityqueue`** — new package: generic min/max heap over
  `container/heap`. `New(less)` instead of implementing five methods.
  `Push` / `Pop` / `Peek` / `Len` / `Clear` / `Drain`.
- **`debounce`** — new package: `Debouncer[T any]` coalesces a burst
  of `Trigger(v)` calls into a single emit on `C()` after a quiet
  window. Latest-wins on slow consumers; injectable `clock.Clock`
  makes it test-deterministic.
- **`lru`** — `Keys()` and `Range(fn)` on both `Cache` and `ShardedCache`,
  skipping expired entries; iteration is MRU-first per shard.
- **`defaultmap`** — `GetOrInit` returns `(V, loaded bool)`; `LoadOrStore`
  for `sync.Map`-style insertion without calling the constructor.
- **`zapctx`** — `PayloadGinMiddleware` logs request/response bodies, status
  and latency, with `WithMaxBody`, `WithSampling`, `WithSkipper` options;
  `PayloadUnaryServerInterceptor` is the gRPC counterpart.
- **`pool`** — `Stats()` exposes workers, in-flight, queued, completed and
  panic counts via cheap atomics.
- **`ttlmap`** — `WithOnExpire` hook fires when an entry is evicted by
  either active access or background sweep.
- **`random`** — `SecureString` uses `crypto/rand` for tokens/salts.
- **`backoff`** — `Backoff.Rand` lets tests inject a deterministic jitter
  source.
- **`caseconv`** — one-step `ToUpperCamel` / `ToLowerSnake` /
  `ToLowerKebab` / etc. that auto-detect the input style.
- **`slogctx`** — new package: `log/slog` counterpart to `zapctx`, including
  `OtelTraceInject`.
- **`errors`** — new package: lightweight stack-trace wrapping with
  `errors.Is`/`As`/`%w` compatibility.
- **`taskgroup`** — new package: `errgroup` plus concurrency limit and
  panic recovery.
- **`ratelimit`** — new package: token bucket and leaky bucket with
  `Allow` / `Wait(ctx)` and injectable clock.
- **`circuitbreaker`** — new package: closed / open / half-open state
  machine with configurable thresholds, cooldown, and `OnStateChange`.
- **`clock`** — new package: `Clock` interface with `Real` and `Mock`
  (`Advance` / `Set`); `Mock.Now` plugs into existing modules' `SetClock`.
- **`httpx`** — new package: `*http.Client` factory combining `backoff`
  retry, `ratelimit`, and `circuitbreaker`; request bodies are buffered
  so they survive retries.
- **`env`** — new package: struct-tag based configuration loader
  (`default=`, `required`) over `os.Getenv` (or a custom map).
- **`redact`** — new package: log-friendly string masking (`Mask`,
  `Email`, `Phone`, `Token`) plus `ZapString` / `SlogString` field helpers.
- **`lru`** — `Resize(n)` changes capacity at runtime, evicting LRU
  entries on shrink.
- **`pool`** — `ScheduleN([]task)` schedules a batch.
- **`taskgroup`** — `Results[T]` collects successful return values from
  concurrent tasks alongside the first error.
- **`errors`** — `Join(errs...)` wraps `stderrors.Join` with a stack;
  `JoinFormatPlusV` prints every joined error on its own line.
- Root `doc.go` summarizes the module's packages.
- `examples/` directory: three runnable programs (`httpserver`,
  `concurrent`, `cache`) demonstrating package combinations.
- CI now runs on Go 1.25.10 and includes `staticcheck`, `gosec`,
  `golangci-lint`, and `govulncheck` jobs.

### Removed

- **`singleflight`** — package removed. It was a thin generic wrapper
  around `golang.org/x/sync/singleflight` with no real added value;
  depend on `x/sync/singleflight` directly. The `examples/cache`
  program now imports it directly. For coalescing requests for
  *different* keys (which singleflight does not do), see the new
  [`batcher`](./batcher) package.

## [v0.1.0] — 2026-05-20

First tagged release. Two waves of internal cleanup are bundled into a
single baseline. APIs are still pre-1.0; expect breaking changes.

### Added

- **`lru`** — generic `Cache[K, V]` with `Peek`, evict hook, and per-entry
  nanosecond-precision expiration. New `ShardedCache` with `Builder.Shards`
  and `BuildSharded` reduces lock contention; `StringHash` helper provided.
- **`itertools`** — `Intersection`, `Contains`, `Chunk`. (Renamed `Reject`
  → `Intersection`.)
- **`defaultmap`** — `Has`, `Len`, `Range`.
- **`pool`** — `Close`, `WithPanicHandler` option, full parameter
  validation. No longer depends on `zap`.
- **`shutdown`** — `Manager` struct, `Wait(ctx)`, `SetLogger`, per-hook
  `WithTimeout`/`WithName` options; panics in hooks are recovered.
- **`zapctx`** — `OtelTraceInject` (OpenTelemetry); `BetterDefault` now
  returns an error; `ZapContext` is exported.
- **`rotatingwriter`** (renamed from `rotating_writer`) — `SizeRotater`
  with optional gzip backups; `Write` is concurrency-safe.
- **`bytesconv`** — modernized to `unsafe.String` / `unsafe.SliceData`.
- **`backoff`** — new package: exponential backoff with jitter,
  `Retry(ctx, ...)`, `Permanent(err)`.
- **`singleflight`** — new package: generic wrapper around
  `golang.org/x/sync/singleflight`.
- **`ttlmap`** — new package: concurrent map with per-entry TTL and an
  optional background sweep.
- Examples (`ExampleXxx`) for lru, itertools, caseconv, defaultmap.
- Fuzz tests for caseconv split/join round-tripping.
- `.golangci.yml` (v2 schema); GitHub Actions CI matrix (Go 1.24, 1.25)
  with codecov upload.
- `lru/benchmark` now compares against `golang-lru/v2` and `ristretto`.

### Changed

- Go directive raised to `1.23` (toolchain `1.25`).
- `caseconv`: replaced deprecated `strings.Title`; constants renamed to
  Go conventions; `AcronymMap` is now hidden behind
  `RegisterAcronym`/`UnregisterAcronym`/`IsAcronym`.
- `random`: uses `math/rand/v2`; concurrent-safe by default.
- `rotatingwriter`: package name no longer contains underscore.

### Removed

- `lru/list.go` (self-rolled doubly linked list) — replaced by
  `container/list`.
- `lru.LRU` type alias and `lru` private type — superseded by
  `Cache[K, V]`.
- `zapctx.OpenTraceInject` and the `go.opencensus.io` dependency.
- `labstack/gommon` indirect dependency.

### Fixed

- `lru.Get`/`Peek` now evict expired entries on access.
- `lru.Clear` / `lru.Size` acquire the lock.
- `random.Sequence.Next` is safe under concurrent first calls.
- `random.String(0, ...)` no longer underflows.
- `rotatingwriter.DoRollover` keeps the previous file on open failure.
- `zapctx.OtelTraceInject` (and the former `OpenTraceInject`) no longer
  panic when the context has no attached logger.
- Test code that previously failed to compile (`string(int)` in lru tests).

[v0.2.0]: https://github.com/liguangsheng/goost/releases/tag/v0.2.0
[v0.1.0]: https://github.com/liguangsheng/goost/releases/tag/v0.1.0
