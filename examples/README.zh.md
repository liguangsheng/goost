# examples

可运行程序，展示这些包如何组合使用。

| 示例 | 演示内容 |
| --- | --- |
| [`httpserver`](./httpserver) | `zapctx` + `zapctxgin` Gin 中间件 + `shutdown` |
| [`concurrent`](./concurrent) | `taskgroup` + `backoff`：在并发限制下重试不稳定任务 |
| [`cache`](./cache) | `lru` + `x/sync/singleflight`：把惊群请求折叠到一次上游调用 |
| [`eventbus`](./eventbus) | `debounce` + `fanout`：嘈杂生产者把突发事件合并成一个事件，并分发给多个订阅者 |

示例位于独立 module 中，避免 demo-only 依赖进入根库 module。运行方式：
`cd examples && go run ./<name>`。
