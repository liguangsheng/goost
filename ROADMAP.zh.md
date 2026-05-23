# 路线图

本路线图把长期规划转成公开检查点。它不是日期承诺。

## v1.0 readiness

v1.0 之前，项目应具备：

- 每个公开包都有最终的 keep/change/remove 决策；
- 对 `httpx`、`zapctx`、`slogctx`、`ttlmap`、`defaultmap` 等短名称或易混名称作出命名决策；
- 每个 v1.0 前仍计划发生的 breaking change 都有迁移说明；
- root、examples、benchmarks 和 optional integrations 的依赖边界稳定；
- 每个公开包和 nested module 都有可编译示例和中英文 README；
- root 与 split-module full gates 能通过 `./scripts/check-release.sh` 跑通；
- 兼容性、弃用、安全、测试和 API 约定都有公开文档。

## v1.0 package audit

本表跟踪公开包当前的 v1.0 方向。`Keep` 表示目前符合项目范围。`Review` 只保留给 v1.0 前仍需要明确包名或 API 形状决策的包。

| Package | Direction | v1.0 note |
| --- | --- | --- |
| `backoff` | Keep | 保持 retry/backoff primitives 小而清晰，并继续感知 context。 |
| `batcher` | Keep | 保持 DataLoader 风格的按 key load coalescing。 |
| `caseconv` | Keep | 保持字符串 case conversion helpers，不扩展到 locale-heavy 范围。 |
| `circuitbreaker` | Keep | 保持为小型下游保护 primitive。 |
| `clock` | Keep | 保持为 deterministic timer 测试抽象。 |
| `debounce` | Keep | 保持 latest-wins quiet-window 语义明确。 |
| `defaultmap` | Keep | 保留该名称；README 和 examples 已明确 lazy construction 语义。 |
| `env` | Keep | 保持小型 struct-tag environment loader，不变成配置框架。 |
| `errors` | Keep | 保持兼容标准 `errors.Is` / `errors.As`。 |
| `fanout` | Keep | 保持进程内 broadcaster 语义和 drop-on-slow 行为明确。 |
| `httpx` | Keep | 保留短名称，并把范围限定为 outbound HTTP client assembly。 |
| `keyedmutex` | Keep | 保持 per-key mutual exclusion 和 idle slot cleanup。 |
| `lru` | Keep | 保持 generic LRU/cache primitives，benchmark 独立维护。 |
| `pool` | Keep | 保持 bounded goroutine pool 行为可观测且 lifecycle-safe。 |
| `priorityqueue` | Keep | 保持 comparator-based heap wrapper 小而清晰。 |
| `random` | Keep | 保持 random string helpers 与 secure generation 边界清楚。 |
| `ratelimit` | Keep | 保持 token/leaky bucket 在 context cancellation 下行为可预测。 |
| `rotatingwriter` | Keep | 保持 file permissions、retention 和 gzip 行为明确。 |
| `shutdown` | Keep | 保持 signal-driven graceful shutdown coordination 范围克制且可移植。 |
| `slogctx` | Keep | 保持为 `zapctx` 的 `log/slog` 对应包，并已记录共享概念。 |
| `taskgroup` | Keep | 保持 errgroup-style concurrency、limit 和 panic recovery。 |
| `ttlmap` | Keep | 保留该名称；close、lazy expiration、sweep 和 purge 语义已有文档和测试。 |
| `zapctx` | Keep | 保留命名，并继续把 framework integrations 留在 nested modules。 |

## 维护复审

每个 release boundary 都要复审：

- package names、package scope 和 keep/change/remove 决策；
- 依赖和 Go toolchain 是否漂移；
- CI action versions 和 cache paths 是否仍然正确；
- 文档链接、examples 和 migration snippets 是否 stale 或损坏；
- 是否存在应弃用或移除的低价值 API；
- backlog ideas 是否仍符合项目范围。

持久决策应进入公开文档、脚本或测试，而不是只留在私有计划里。

复审后，应排序下一轮 10 到 20 个具体执行切片。每个切片都要有负责的改动面、预期制品和验证命令。宽泛想法必须先通过项目政策中的新增准入标准，才能进入 roadmap。

## next execution slices

11-43 切片已完成。下表是已完成的队列。

| Slice | Surface | Artifact | Validation |
| --- | --- | --- | --- |
| 32 | v0.4.0 发布 | CHANGELOG 定稿，tag 待确认 | `./scripts/check-release.sh` |
| 33 | godoc 渲染审计 | 所有 exported symbols 有 doc comment | `go doc ./...` 并检查空白项 |
| 34 | 测试覆盖率基线 | 各包覆盖率已记录到 TESTING.md（总计 91.7%） | `go test -coverprofile=coverage.out ./...` |
| 35 | 错误链审计 | 所有 sentinel 添加 errors.Is 测试 | `./scripts/check-root.sh --quick` |
| 36 | consumer contract tests | 编译时接口检查和 API 表面测试 | `go test ./... -run Consumer` |
| 37 | `httpx` hook panic 安全性 | Hook panic recovery 已验证 | `go test -race ./httpx` |
| 38 | 构造函数验证测试 | 所有构造函数的无效输入测试 | `./scripts/check-root.sh --quick` |
| 39 | 文档链接检查器 | check-doc-links.sh 集成到 release gate | `./scripts/check-release.sh` |
| 40 | release gate 刷新 | 全部切片后 full gate 通过 | `./scripts/check-release.sh` |
| 41 | `sync` 原语审计 | WaitGroup/Once/Map 使用已验证正确 | `go vet ./...` |
| 42 | examples 冒烟 gate | examples 的 split-module gate 通过 | `./scripts/check-split-modules.sh --quick --module ./examples` |
| 43 | Go 1.26 兼容性探针 | 允许失败的 root smoke job 使用 Go 1.26.3 运行 | GitHub Actions `go-next-root-smoke` |

下一轮切片将在 v0.4.0 发布后规划。
