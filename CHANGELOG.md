# Changelog

All notable changes to this project will be documented in this file.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

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
- CI now runs `staticcheck` and `gosec` jobs in addition to
  `golangci-lint`.

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

[v0.1.0]: https://github.com/liguangsheng/goost/releases/tag/v0.1.0
