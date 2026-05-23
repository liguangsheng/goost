# Testing Strategy

This project keeps validation split by cost and risk.

## Daily Gates

Root-module changes:

```sh
./scripts/check-root.sh --quick
```

One nested module:

```sh
./scripts/check-split-modules.sh --quick --module <path>
```

## Release Gate

Before publishing a release:

```sh
./scripts/check-release.sh
```

This runs the full root and split-module gates, including heavier race,
security, vulnerability, and split-module checks.

The release gate also runs `./scripts/check-scripts.sh`, which checks script
syntax, executable bits, help output, nested-module discovery, and CI cache path
alignment.

CI cache path alignment is checked by `scripts/check-ci-cache-paths.sh`. The
script discovers every repository `go.sum` outside `.git` and `.agents`, parses
both block and single-line `cache-dependency-path` entries in
`.github/workflows/ci.yml`, and fails when the two sets drift. This keeps root,
examples, benchmarks, and optional integration modules in the same cache policy.

## Cross-Platform Smoke

CI includes a Windows root smoke job that runs `go test ./...` without the heavy
analysis toolchain. The full release gate still runs on Ubuntu, but the Windows
job catches path, permission, signal, timer, and HTTP assumptions that are easy
to miss in Linux-only development.

## Fuzz Tests

Fuzz tests cover input-heavy code such as `caseconv`. Run them intentionally,
then promote useful discoveries into ordinary regression tests:

```sh
go test ./caseconv -run=^$ -fuzz=Fuzz -fuzztime=30s
```

## Benchmarks

Benchmarks are local performance evidence, not correctness checks. Run LRU
benchmarks from the benchmark nested module:

```sh
cd lru/benchmark && go test -bench=. ./...
```

## Stress and Race Tests

Stress tests live beside concurrency-heavy packages and are included in normal
package tests when they are stable. Use the full root gate before release to run
the root packages with `-race`.

For a focused stress pass, run:

```sh
./scripts/check-stress.sh --quick
```

For the same stress-focused packages under the race detector, run:

```sh
./scripts/check-stress.sh --race
```

Current stress-focused coverage:

| Package | Why it is covered |
| --- | --- |
| `batcher` | Coalesces concurrent callers into shared windows, so stress coverage exercises queueing, cancellation, and in-flight batch accounting. |
| `fanout` | Drops instead of blocking slow subscribers, so stress coverage exercises publish pressure, subscriber close paths, and drop counters. |
| `keyedmutex` | Coordinates per-key lock slots, so stress coverage exercises contention, unlock ordering, and idle slot cleanup. |
| `pool` | Owns worker goroutines and optional queues, so stress coverage exercises queue pressure, panic recovery, shutdown, and stats. |
| `ttlmap` | Owns expiration state and an optional sweep goroutine, so stress coverage exercises timer cleanup, lazy expiration, purge, and close behavior. |

Keep long-running ad hoc stress loops outside the supported gate scripts until
they are stable enough for repeatable local use.

Concurrency-heavy packages should cover cancellation, shutdown, queue pressure,
and timer cleanup with deterministic synchronization where possible. Prefer
fake clocks, channels, and explicit handshakes over sleeps. When a test must use
real time, keep the timeout generous and the assertion narrow.

## Observability and Lifecycle Tests

Packages that expose `Stats`, `Snapshot`, callbacks, or hooks should test both
normal and edge states: empty, active, error, canceled, and closed. Snapshot
tests should assert the documented meaning of gauges, counters, configuration
values, and derived values.

Types that own goroutines, timers, files, or network resources should have tests
for their documented release path. `Close`, `Stop`, and `Wait` behavior should be
covered for repeated calls when the API promises idempotency.

## Test Style

Prefer table-driven tests for pure input/output behavior and small validation
matrices. Keep direct scenario tests for concurrency, lifecycle, ordering,
timing, and smoke checks where a table would hide the sequence being exercised.

The repository intentionally mixes standard-library assertions with
`testify/assert` and a small amount of `testify/require`: use `assert` for most
package-level expectations, `require` for setup preconditions that make the rest
of a test meaningless, and `t.Fatal`/`t.Errorf` for smoke tests, fuzz tests, and
select-based concurrency assertions. Helper names should describe their role
directly, such as `testResponse`, `newCapturingLogger`, or `fakeLimiter`.

## Structured Logging Tests

Logging tests should assert field names and field values, not just that some
text was written. Use stable in-memory loggers such as `slog.Handler` test
doubles or zap's observer core so assertions are independent of timestamps,
random IDs, and formatter changes.

HTTP and payload logging tests must assert that query strings, request bodies,
tokens, passwords, and other sensitive values are not emitted unless the package
explicitly documents that behavior.

## Coverage Baseline

The current per-package test coverage baseline (generated with `go test
-coverprofile=coverage.out -covermode=atomic ./...`):

| Package | Coverage |
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
| `rotatingwriter` | 84.8% |
| `shutdown` | 91.7% |
| `slogctx` | 94.7% |
| `taskgroup` | 96.8% |
| `ttlmap` | 100.0% |
| `zapctx` | 88.5% |

**Total: 91.7%**

Packages below 80% should be evaluated for additional test coverage. No package
is currently below that threshold.

The full root gate (`./scripts/check-root.sh --full`) outputs a coverage summary.
The baseline is recorded here for tracking; there is no hard CI threshold, but
coverage should not regress without justification.
