# goost

A small collection of Go utilities the author keeps reaching for.

## Install

```sh
go get github.com/liguangsheng/goost
```

## Packages

| Package | Purpose |
| --- | --- |
| [`bytesconv`](./bytesconv) | Allocation-free `string`/`[]byte` conversion (read-only). |
| [`caseconv`](./caseconv) | Camel/snake/kebab/pascal case split & join helpers. |
| [`defaultmap`](./defaultmap) | Concurrent map that lazy-constructs values for missing keys. |
| [`itertools`](./itertools) | Generic slice helpers: `Map`, `Filter`, `Reduce`, `Chunk`, etc. |
| [`lru`](./lru) | Generic LRU cache with optional per-entry expiration. |
| [`pool`](./pool) | Bounded goroutine pool with optional queue and panic recovery. |
| [`random`](./random) | Concurrency-safe random string generator. |
| [`rotating_writer`](./rotating_writer) | `io.Writer` that rotates the backing file (e.g. daily). |
| [`shutdown`](./shutdown) | Signal-driven graceful shutdown coordinator. |
| [`zapctx`](./zapctx) | Carry a `*zap.Logger` and structured fields through `context.Context`. |

All packages are independent; depend on what you need.

## Stability

The module is still pre-1.0. APIs may change between minor versions.

## License

MIT — see [LICENSE](./LICENSE).
