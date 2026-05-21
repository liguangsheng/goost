# examples

Runnable programs that show how the packages fit together.

| Example | Demonstrates |
| --- | --- |
| [`httpserver`](./httpserver) | `zapctx` + `zapctxgin` Gin middleware + `shutdown` |
| [`concurrent`](./concurrent) | `taskgroup` + `backoff` retrying flaky tasks under a concurrency limit |
| [`cache`](./cache) | `lru` + `x/sync/singleflight` collapsing a thundering herd onto one upstream call |
| [`eventbus`](./eventbus) | `debounce` + `fanout`: a noisy producer collapses bursts into one event delivered to many subscribers |

Examples live in their own module so demo-only dependencies stay out of the
root library module. Run with `cd examples && go run ./<name>`.
