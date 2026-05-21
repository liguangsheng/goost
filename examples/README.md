# examples

Runnable programs that show how the packages fit together.

| Example | Demonstrates |
| --- | --- |
| [`httpserver`](./httpserver) | `zapctx` + `zapctxgin` Gin middleware + `zapctxotel` trace hook + `shutdown` |
| [`concurrent`](./concurrent) | `taskgroup` + `backoff` retrying flaky tasks under a concurrency limit |
| [`cache`](./cache) | `lru` + `x/sync/singleflight` collapsing a thundering herd onto one upstream call |
| [`eventbus`](./eventbus) | `debounce` + `fanout`: a noisy producer collapses bursts into one event delivered to many subscribers |

Each is a `package main`; run with `go run ./examples/<name>`.
