# goost

A small collection of Go utilities the author keeps reaching for.

## Install

```sh
go get github.com/liguangsheng/goost
```

## Packages

| Package | Purpose |
| --- | --- |
| [`backoff`](./backoff) | Exponential backoff with jitter and a context-aware `Retry`. |
| [`bytesconv`](./bytesconv) | Allocation-free `string`/`[]byte` conversion (read-only). |
| [`caseconv`](./caseconv) | Camel/snake/kebab/pascal case split & join helpers. |
| [`circuitbreaker`](./circuitbreaker) | Three-state breaker (closed/open/half-open) for downstream call protection. |
| [`defaultmap`](./defaultmap) | Concurrent map that lazy-constructs values for missing keys. |
| [`errors`](./errors) | `errors` with stack traces; `errors.Is`/`As`/`%w` compatible. |
| [`itertools`](./itertools) | Generic slice helpers: `Map`, `Filter`, `Reduce`, `Chunk`, etc. |
| [`lru`](./lru) | Generic LRU cache with optional per-entry expiration; sharded variant; `Keys`/`Range`. |
| [`pool`](./pool) | Bounded goroutine pool with optional queue, panic handler, and `Stats`. |
| [`random`](./random) | Concurrency-safe random string generator; `SecureString` via `crypto/rand`. |
| [`ratelimit`](./ratelimit) | Token bucket and leaky bucket rate limiters. |
| [`rotatingwriter`](./rotatingwriter) | `io.Writer` that rotates the backing file (daily or size-based, optional gzip). |
| [`shutdown`](./shutdown) | Signal-driven graceful shutdown coordinator (per-hook timeouts). |
| [`singleflight`](./singleflight) | Generic wrapper around `x/sync/singleflight`. |
| [`slogctx`](./slogctx) | Carry a `*slog.Logger` and attrs through `context.Context`. |
| [`taskgroup`](./taskgroup) | `errgroup` + concurrency limit + panic recovery. |
| [`ttlmap`](./ttlmap) | Concurrent map with per-entry expiration and background sweep. |
| [`zapctx`](./zapctx) | Carry a `*zap.Logger` and structured fields through `context.Context`. |

Runnable end-to-end programs live in [`examples/`](./examples).

All packages are independent; depend on what you need.

## Stability

The module is still pre-1.0. APIs may change between minor versions.
See [CHANGELOG.md](./CHANGELOG.md) for release history.

## License

MIT — see [LICENSE](./LICENSE).
