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

These slices are the current public queue after the first v1.0 package audit
and release dry-run. Each slice is intentionally small enough to land with a
clear artifact and validation command.

| Slice | Surface | Artifact | Validation |
| --- | --- | --- | --- |
| 11 | observability wording | Normalize Stats/Snapshot wording for `batcher`, `fanout`, `pool`, and `ratelimit` | `./scripts/check-root.sh --quick` |
| 12 | stress coverage notes | Document which packages are covered by `scripts/check-stress.sh` and why | `go test .` |
| 13 | `httpx` body replay fixtures | Add compile/runtime coverage for replayable and non-replayable request bodies | `go test ./httpx` |
| 14 | `ratelimit` cancellation | Tighten docs and tests for context cancellation while waiting | `go test ./ratelimit` |
| 15 | `pool` shutdown semantics | Clarify submit, close, wait, and stats behavior around shutdown | `go test ./pool` |
| 16 | `taskgroup` panic behavior | Align README, example, and tests for panic recovery and error propagation | `go test ./taskgroup` |
| 17 | examples smoke output | Keep runnable example outputs deterministic and documented | `./scripts/check-split-modules.sh --quick --module ./examples` |
| 18 | CI cache drift guard | Recheck nested module discovery against GitHub Actions cache paths | `go test .` |
| 19 | release docs parity | Re-audit changelog, migration guide, and localized links after the next changes | `./scripts/check-root.sh --quick` |
| 20 | release gate repeatability | Re-run and record `./scripts/check-release.sh` after slices 11-19 | `./scripts/check-release.sh` |
