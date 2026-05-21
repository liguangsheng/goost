# examples

可运行程序，展示这些包如何组合使用。

| 示例 | 演示内容 |
| --- | --- |
| [`httpserver`](./httpserver) | `zapctx` + `zapctxgin` Gin 中间件 + `zapctxotel` trace hook + `shutdown` |
| [`concurrent`](./concurrent) | `taskgroup` + `backoff`：在并发限制下重试不稳定任务 |
| [`cache`](./cache) | `lru` + `x/sync/singleflight`：把惊群请求折叠到一次上游调用 |
| [`eventbus`](./eventbus) | `debounce` + `fanout`：嘈杂生产者把突发事件合并成一个事件，并分发给多个订阅者 |

每个示例都是 `package main`；使用 `go run ./examples/<name>` 运行。
