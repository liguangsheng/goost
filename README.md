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
| [`defaultmap`](./defaultmap) | Concurrent map that lazy-constructs values for missing keys. |
| [`itertools`](./itertools) | Generic slice helpers: `Map`, `Filter`, `Reduce`, `Chunk`, etc. |
| [`lru`](./lru) | Generic LRU cache with optional per-entry expiration; sharded variant. |
| [`pool`](./pool) | Bounded goroutine pool with optional queue and panic handler. |
| [`random`](./random) | Concurrency-safe random string generator. |
| [`rotatingwriter`](./rotatingwriter) | `io.Writer` that rotates the backing file (daily or size-based, optional gzip). |
| [`shutdown`](./shutdown) | Signal-driven graceful shutdown coordinator (per-hook timeouts). |
| [`singleflight`](./singleflight) | Generic wrapper around `x/sync/singleflight`. |
| [`ttlmap`](./ttlmap) | Concurrent map with per-entry expiration and background sweep. |
| [`zapctx`](./zapctx) | Carry a `*zap.Logger` and structured fields through `context.Context`. |

All packages are independent; depend on what you need.

## Stability

The module is still pre-1.0. APIs may change between minor versions.

## License

MIT — see [LICENSE](./LICENSE).
