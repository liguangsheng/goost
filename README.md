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
| [`batcher`](./batcher) | DataLoader-style coalescing of concurrent per-key loads into one batch call. |
| [`bytesconv`](./bytesconv) | Allocation-free `string`/`[]byte` conversion (read-only). |
| [`caseconv`](./caseconv) | Camel/snake/kebab/pascal case split & join helpers. |
| [`circuitbreaker`](./circuitbreaker) | Three-state breaker (closed/open/half-open) for downstream call protection. |
| [`clock`](./clock) | `Clock` abstraction with `Real` and `Mock` for deterministic time tests. |
| [`debounce`](./debounce) | Coalesce a burst of `Trigger(v)` calls into one emit after a quiet window; latest-wins. |
| [`defaultmap`](./defaultmap) | Concurrent map that lazy-constructs values for missing keys. |
| [`env`](./env) | Struct-tag configuration loader from environment variables. |
| [`errors`](./errors) | `errors` with stack traces; `Join`; `Recover` (defer panic→error); `Is`/`As`/`%w` compatible. |
| [`fanout`](./fanout) | In-process broadcaster: one publisher, many subscribers; drop-on-slow. |
| [`httpx`](./httpx) | `*http.Client` with retry + ratelimit + circuit breaker. |
| [`itertools`](./itertools) | Generic slice helpers: `Map`, `Filter`, `Reduce`, `Chunk`, etc. |
| [`keyedmutex`](./keyedmutex) | Per-key mutex: parallel across keys, serial per key; slots GC when idle. |
| [`lru`](./lru) | Generic LRU cache with optional per-entry expiration; sharded variant; `Keys`/`Range`/`Resize`. |
| [`pool`](./pool) | Bounded goroutine pool with optional queue, panic handler, and `Stats`. |
| [`priorityqueue`](./priorityqueue) | Generic min/max heap over `container/heap` — comparator instead of five methods. |
| [`redact`](./redact) | String masking for logs (`Email`/`Phone`/`Token`/`Mask`). |
| [`random`](./random) | Concurrency-safe random string generator; `SecureString` via `crypto/rand`. |
| [`ratelimit`](./ratelimit) | Token bucket and leaky bucket rate limiters. |
| [`rotatingwriter`](./rotatingwriter) | `io.Writer` that rotates the backing file (daily or size-based, optional gzip). |
| [`shutdown`](./shutdown) | Signal-driven graceful shutdown coordinator (per-hook timeouts). |
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
