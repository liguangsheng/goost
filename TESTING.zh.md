# 测试策略

本项目按成本和风险拆分验证路径。

## 日常 gate

Root module 改动：

```sh
./scripts/check-root.sh --quick
```

单个 nested module：

```sh
./scripts/check-split-modules.sh --quick --module <path>
```

## 发布 gate

发布前运行：

```sh
./scripts/check-release.sh
```

它会运行 root 与 split-module 的 full gates，包含较重的 race、安全、漏洞和 split-module 检查。

发布 gate 也会运行 `./scripts/check-scripts.sh`，检查脚本语法、可执行位、help 输出、nested-module discovery 和 CI cache path 是否一致。

CI cache path 一致性由 `scripts/check-ci-cache-paths.sh` 检查。该脚本会发现
`.git` 和 `.agents` 之外的所有仓库 `go.sum`，解析 `.github/workflows/ci.yml`
中的 block 和 single-line `cache-dependency-path` 写法，并在两边集合漂移时失败。
这能保证 root、examples、benchmarks 和 optional integration modules 使用同一套 cache policy。

## 跨平台 smoke

CI 包含一个 Windows root smoke job，会运行 `go test ./...`，但不安装较重的分析工具链。完整 release gate 仍在 Ubuntu 上运行；Windows job 用来尽早发现 Linux-only 开发中容易漏掉的 path、permission、signal、timer 和 HTTP 假设。

## Fuzz tests

Fuzz tests 用于覆盖输入边界较多的代码，例如 `caseconv`。请有意识地运行它们，并把有价值的发现固化成普通回归测试：

```sh
go test ./caseconv -run=^$ -fuzz=Fuzz -fuzztime=30s
```

## Benchmarks

Benchmarks 是本地性能证据，不是正确性检查。LRU benchmark 请在 benchmark nested module 中运行：

```sh
cd lru/benchmark && go test -bench=. ./...
```

## Stress 与 race tests

Stress tests 放在并发核心包旁边；只要稳定，就纳入普通 package tests。发布前使用 full root gate 以 `-race` 运行 root packages。

聚焦运行 stress tests：

```sh
./scripts/check-stress.sh --quick
```

对同一组 stress-focused packages 运行 race detector：

```sh
./scripts/check-stress.sh --race
```

当前 stress-focused 覆盖范围：

| Package | 覆盖原因 |
| --- | --- |
| `batcher` | 会把并发调用者合并到共享窗口，因此 stress coverage 覆盖排队、cancellation 和 in-flight batch 统计。 |
| `fanout` | 对慢订阅者选择丢弃而不是阻塞，因此 stress coverage 覆盖 publish pressure、subscriber close paths 和 drop counters。 |
| `keyedmutex` | 协调 per-key lock slots，因此 stress coverage 覆盖 contention、unlock ordering 和 idle slot cleanup。 |
| `pool` | 拥有 worker goroutines 和可选队列，因此 stress coverage 覆盖 queue pressure、panic recovery、shutdown 和 stats。 |
| `ttlmap` | 拥有 expiration state 和可选 sweep goroutine，因此 stress coverage 覆盖 timer cleanup、lazy expiration、purge 和 close behavior。 |

长时间 ad hoc stress loops 在足够稳定、可以重复本地运行之前，不应进入受支持的 gate scripts。

并发核心包应覆盖 cancellation、shutdown、queue pressure 和 timer cleanup，尽量使用确定性的同步方式。优先使用 fake clock、channel 和明确的握手信号，不优先使用 sleep。必须使用真实时间时，timeout 要留足，断言要收窄。

## 观测与生命周期测试

暴露 `Stats`、`Snapshot`、callback 或 hook 的包，应覆盖正常状态和边界状态：empty、active、error、canceled、closed。Snapshot 测试应断言文档中定义的 gauge、counter、配置值和派生值语义。

拥有 goroutine、timer、文件或网络资源的类型，应为文档中的释放路径补测试。如果 API 承诺 `Close`、`Stop` 或 `Wait` 可重复调用，也要覆盖重复调用行为。

## 测试风格

纯输入/输出行为和小型验证矩阵优先使用 table-driven tests。并发、生命周期、顺序、计时和 smoke checks 可以保留直接场景测试；这类测试改成表格反而容易隐藏被验证的执行顺序。

本仓库有意混用标准库断言、`testify/assert` 和少量 `testify/require`：package-level expectations 默认用 `assert`；测试前置条件失败后其余断言没有意义时用 `require`；smoke tests、fuzz tests 和 select-based concurrency assertions 使用 `t.Fatal`/`t.Errorf`。Helper 名称应直接说明角色，例如 `testResponse`、`newCapturingLogger` 或 `fakeLimiter`。

## 结构化日志测试

日志测试应断言字段名和字段值，而不是只检查写出过某段文本。优先使用稳定的内存 logger，例如 `slog.Handler` test double 或 zap observer core，让断言不依赖时间戳、随机 ID 或 formatter 变化。

HTTP 和 payload logging 测试必须断言 query string、request body、token、password 和其他敏感值不会被写出，除非对应包明确把这种行为写进文档。

## 覆盖率基线

当前各包测试覆盖率基线（通过 `go test -coverprofile=coverage.out -covermode=atomic ./...` 生成）：

| Package | 覆盖率 |
| --- | --- |
| `backoff` | 86.0% |
| `batcher` | 93.8% |
| `caseconv` | 83.2% |
| `circuitbreaker` | 96.0% |
| `clock` | 91.5% |
| `debounce` | 95.3% |
| `defaultmap` | 95.0% |
| `env` | 89.2% |
| `errors` | 88.8% |
| `fanout` | 98.0% |
| `httpx` | 96.5% |
| `keyedmutex` | 95.7% |
| `lru` | 86.1% |
| `pool` | 95.9% |
| `priorityqueue` | 100.0% |
| `random` | 97.6% |
| `ratelimit` | 92.8% |
| `rotatingwriter` | 85.3% |
| `shutdown` | 91.7% |
| `slogctx` | 94.7% |
| `taskgroup` | 96.8% |
| `ttlmap` | 100.0% |
| `zapctx` | 88.5% |

**总计：91.8%**

低于 80% 的包应评估是否需要补充测试覆盖。当前没有包低于该阈值。

Full root gate（`./scripts/check-root.sh --full`）会输出覆盖率摘要。此基线记录在此用于跟踪；没有硬性 CI 门槛，但覆盖率不应无理由地退化。
