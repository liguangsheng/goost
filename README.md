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
| [`caseconv`](./caseconv) | Camel/snake/kebab/pascal case split & join helpers. |
| [`circuitbreaker`](./circuitbreaker) | Three-state breaker (closed/open/half-open) for downstream call protection. |
| [`clock`](./clock) | `Clock` abstraction with `Real` and `Mock` for deterministic time tests. |
| [`debounce`](./debounce) | Coalesce a burst of `Trigger(v)` calls into one emit after a quiet window; latest-wins. |
| [`defaultmap`](./defaultmap) | Concurrent map that lazy-constructs values for missing keys. |
| [`env`](./env) | Struct-tag configuration loader from environment variables. |
| [`errors`](./errors) | `errors` with stack traces; `Join`; `Recover` (defer panic→error); `Is`/`As`/`%w` compatible. |
| [`fanout`](./fanout) | In-process broadcaster: one publisher, many subscribers; drop-on-slow. |
| [`httpx`](./httpx) | `*http.Client` with retry, ratelimit, circuit breaker, and request logging. |
| [`keyedmutex`](./keyedmutex) | Per-key mutex: parallel across keys, serial per key; slots GC when idle. |
| [`lru`](./lru) | Generic LRU cache with optional per-entry expiration; sharded variant; `Keys`/`Range`/`Resize`. |
| [`pool`](./pool) | Bounded goroutine pool with optional queue, panic handler, and `Stats`. |
| [`priorityqueue`](./priorityqueue) | Generic min/max heap over `container/heap` — comparator instead of five methods. |
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

## Optional Integration Modules

| Module | Purpose |
| --- | --- |
| [`zapctx/zapctxgin`](./zapctx/zapctxgin) | Gin middleware and HTTP payload logging for `zapctx`. |
| [`zapctx/zapctxgrpc`](./zapctx/zapctxgrpc) | gRPC interceptors and payload logging for `zapctx`. |

## Stability

The module is still pre-1.0. APIs may change between minor versions.
See [CHANGELOG.md](./CHANGELOG.md) for release history.
See [MIGRATION.md](./MIGRATION.md) for minor-version migration notes.

## Development Checks

The primary supported Go toolchain version is `1.25.10`, matching every
`go.mod` file and the CI `GO_VERSION` setting. CI also runs an allow-failure
root smoke check on Go `1.26.3` to catch upcoming compatibility issues early.

For day-to-day root-module changes, run `./scripts/check-root.sh --quick`.
Nested modules are separate Go modules, so root checks do not traverse them.
The root gate also verifies that every README-listed public package has a
compiled `Example` test; runnable programs in `examples/` are covered by the
split-module gate instead.

CI installs the required analysis tools through
`./scripts/install-ci-tools.sh`, with tool versions controlled by the workflow
environment.

Contribution guidance lives in [CONTRIBUTING.md](./CONTRIBUTING.md).
Long-lived scope, terminology, and deprecation rules live in
[PROJECT_POLICY.md](./PROJECT_POLICY.md).
Security-sensitive logging and file-permission guidance lives in
[SECURITY.md](./SECURITY.md).
Public API shape, zero-value, generic, and error conventions live in
[API_CONVENTIONS.md](./API_CONVENTIONS.md).
Testing, fuzzing, benchmark, stress, and release-gate guidance lives in
[TESTING.md](./TESTING.md).

Before publishing a release, run:

```sh
./scripts/check-release.sh
```

For day-to-day nested-module changes, run the split-module gate against the
module you touched:

```sh
./scripts/check-split-modules.sh --quick --module zapctx/zapctxgin
```

## License

MIT — see [LICENSE](./LICENSE).
