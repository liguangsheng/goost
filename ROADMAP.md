# Roadmap

This roadmap turns long-term planning into public checkpoints. It is not a
promise of dates.

## v1.0 Readiness

Before v1.0, the project should have:

- a final keep/change/remove decision for every public package,
- a naming decision for packages with short or ambiguous names such as `httpx`,
  `zapctx`, `slogctx`, `ttlmap`, and `defaultmap`,
- migration guidance for every breaking change still planned before v1.0,
- stable dependency boundaries for root, examples, benchmarks, and optional
  integrations,
- compiled examples and bilingual README files for every public package and
  nested module,
- full root and split-module gates passing through `./scripts/check-release.sh`,
- documented compatibility, deprecation, security, testing, and API conventions.

## v1.0 Package Audit

This table tracks the current v1.0 direction for public packages. `Keep` means
the package fits the project scope today. `Review` is reserved for a package
whose name or API shape still needs an explicit decision before v1.0.

| Package | Direction | v1.0 note |
| --- | --- | --- |
| `backoff` | Keep | Keep retry/backoff primitives small and context-aware. |
| `batcher` | Keep | Keep DataLoader-style batching focused on per-key load coalescing. |
| `caseconv` | Keep | Keep as string case conversion helpers; avoid locale-heavy scope. |
| `circuitbreaker` | Keep | Keep as a small downstream protection primitive. |
| `clock` | Keep | Keep as test-time abstraction for deterministic timers. |
| `debounce` | Keep | Keep latest-wins quiet-window semantics explicit. |
| `defaultmap` | Keep | Keep the name; README and examples now spell out lazy construction semantics. |
| `env` | Keep | Keep as a small struct-tag environment loader, not a config framework. |
| `errors` | Keep | Keep compatibility with standard `errors.Is` / `errors.As`. |
| `fanout` | Keep | Keep in-process broadcaster semantics and drop-on-slow behavior explicit. |
| `httpx` | Keep | Keep the short name while limiting scope to outbound HTTP client assembly. |
| `keyedmutex` | Keep | Keep per-key mutual exclusion with idle slot cleanup. |
| `lru` | Keep | Keep generic LRU/cache primitives and benchmark separately. |
| `pool` | Keep | Keep bounded goroutine pool behavior observable and lifecycle-safe. |
| `priorityqueue` | Keep | Keep comparator-based heap wrapper small. |
| `random` | Keep | Keep random string helpers and secure generation clearly separated. |
| `ratelimit` | Keep | Keep token and leaky bucket limiters predictable under context cancellation. |
| `rotatingwriter` | Keep | Keep file permissions, retention, and gzip behavior explicit. |
| `shutdown` | Keep | Keep signal-driven graceful shutdown coordination scoped and portable. |
| `slogctx` | Keep | Keep as the `log/slog` counterpart to `zapctx`, with shared concepts documented. |
| `taskgroup` | Keep | Keep errgroup-style concurrency with limit and panic recovery. |
| `ttlmap` | Keep | Keep the name; close, lazy expiration, sweep, and purge semantics are documented and tested. |
| `zapctx` | Keep | Keep naming and keep framework integrations in nested modules. |

## Maintenance Review

At each release boundary, review:

- package names, package scope, and keep/change/remove decisions,
- dependency and Go toolchain drift,
- CI action versions and cache paths,
- stale docs, broken links, examples, and migration snippets,
- low-value APIs that should be deprecated or removed,
- backlog ideas that no longer match the project scope.

Move durable decisions into public docs, scripts, or tests instead of leaving
them only in private planning notes.

After the review, rank the next 10 to 20 concrete execution slices. Each slice
should have an owner surface, expected artifact, and validation command. Do not
promote broad ideas into the roadmap until they pass the project policy's
addition criteria.

## Next Execution Slices

Slices 11-20 are complete. The table below is the current queue for the next
round.

| Slice | Surface | Artifact | Validation |
| --- | --- | --- | --- |
| 21 | `lru` observability wording | Normalize Snapshot wording for `lru` and `ShardedCache` | `./scripts/check-root.sh --quick` |
| 22 | `circuitbreaker` observability wording | Normalize Snapshot wording for consecutive counters and timestamps | `./scripts/check-root.sh --quick` |
| 23 | `ratelimit` stress tests | Add concurrent Wait stress tests for Bucket and Leaky | `go test -race ./ratelimit` |
| 24 | `debounce` concurrent tests | Add concurrent Trigger + Stop stress test | `go test -race ./debounce` |
| 25 | `pool` cancellation tests | Add Schedule/ScheduleTimeout with pre-cancelled context | `go test -race ./pool` |
| 26 | `fanout` concurrent edge tests | Add subscribe/close during concurrent publish tests | `go test -race ./fanout` |
| 27 | package doc parity | Add doc.go for `caseconv` and `random` | `go vet ./caseconv ./random` |
| 28 | stress gate refresh | Re-run and record `./scripts/check-stress.sh` after slices 21-27 | `./scripts/check-stress.sh` |
| 29 | release gate refresh | Re-run and record `./scripts/check-release.sh` after slices 21-28 | `./scripts/check-release.sh` |
| 30 | `debounce` race fix | Fix close-channel race between `emit()` and `Stop()` | `go test -race ./debounce` |
| 31 | examples module gate | Re-run and record split-module gate for examples | `./scripts/check-split-modules.sh --quick --module ./examples` |
