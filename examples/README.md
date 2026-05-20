# examples

Runnable programs that show how the packages fit together.

| Example | Demonstrates |
| --- | --- |
| [`httpserver`](./httpserver) | `zapctx` + `gin` middleware + `OtelTraceInject` + `shutdown` |
| [`concurrent`](./concurrent) | `taskgroup` + `backoff` retrying flaky tasks under a concurrency limit |
| [`cache`](./cache) | `lru` + `singleflight` collapsing a thundering herd onto one upstream call |

Each is a `package main`; run with `go run ./examples/<name>`.
