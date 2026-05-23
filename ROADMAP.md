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

Slices 11-42 are complete. The table below is the completed queue.

| Slice | Surface | Artifact | Validation |
| --- | --- | --- | --- |
| 32 | v0.4.0 release | CHANGELOG finalized, tag pending approval | `./scripts/check-release.sh` |
| 33 | godoc rendering audit | All exported symbols have doc comments | `go doc ./...` and check for empty entries |
| 34 | test coverage baseline | Per-package coverage recorded in TESTING.md (total 91.3%) | `go test -coverprofile=coverage.out ./...` |
| 35 | error chain audit | errors.Is tests for all sentinels | `./scripts/check-root.sh --quick` |
| 36 | consumer contract tests | Compile-time interface checks and API surface tests | `go test ./... -run Consumer` |
| 37 | `httpx` hook panic safety | Hook panic recovery verified | `go test -race ./httpx` |
| 38 | constructor validation tests | Invalid-input tests for all constructors | `./scripts/check-root.sh --quick` |
| 39 | doc link checker | check-doc-links.sh in release gate | `./scripts/check-release.sh` |
| 40 | release gate refresh | Full gate passes after all slices | `./scripts/check-release.sh` |
| 41 | `sync` primitive audit | WaitGroup/Once/Map usage verified correct | `go vet ./...` |
| 42 | examples smoke gate | Split-module gate passes for examples | `./scripts/check-split-modules.sh --quick --module ./examples` |

The next slices will be planned after the v0.4.0 release.
