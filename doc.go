// Package goost is a meta-package; its sub-packages collect small, generic
// Go utilities the author uses across projects.
//
// The packages are independent — depend on whichever you need:
//
//   - backoff:         exponential backoff with jitter and a Retry helper
//   - bytesconv:       allocation-free string/[]byte conversion (read-only)
//   - caseconv:        camel/snake/kebab/pascal case split & join
//   - circuitbreaker:  three-state breaker (closed / open / half-open)
//   - defaultmap:      concurrent map that lazy-constructs missing values
//   - errors:          stack-tracing wrap on top of stdlib errors
//   - itertools:       generic slice helpers (Map/Filter/Reduce/...)
//   - lru:             generic LRU cache; optional TTL; sharded variant
//   - pool:            bounded goroutine pool with panic handler & Stats
//   - random:          random strings; SecureString uses crypto/rand
//   - ratelimit:       token bucket and leaky bucket limiters
//   - rotatingwriter:  io.Writer with daily or size-based rotation
//   - shutdown:        signal-driven graceful shutdown coordinator
//   - singleflight:    generic wrapper around x/sync/singleflight
//   - slogctx:         log/slog companion to zapctx
//   - taskgroup:       errgroup + concurrency limit + panic recovery
//   - ttlmap:          concurrent map with per-entry TTL and OnExpire
//   - zapctx:          *zap.Logger and structured fields via context.Context
//
// See CHANGELOG.md for release notes and the examples/ directory for
// end-to-end programs that combine several packages.
package goost
