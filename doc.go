// Package goost is a meta-package; its sub-packages collect small, generic
// Go utilities the author uses across projects.
//
// The packages are independent — depend on whichever you need:
//
//   - backoff:         exponential backoff with jitter and a Retry helper
//   - batcher:         DataLoader-style per-key request coalescing
//   - caseconv:        camel/snake/kebab/pascal case split & join
//   - circuitbreaker:  three-state breaker (closed / open / half-open)
//   - clock:           Clock abstraction with Real and Mock for tests
//   - debounce:        coalesce bursts into one emit after a quiet window
//   - defaultmap:      concurrent map that lazy-constructs missing values
//   - env:             struct-tag configuration loader from environment vars
//   - errors:          stack-tracing wrap on top of stdlib errors; Recover
//   - fanout:          in-process broadcaster with drop-on-slow backpressure
//   - httpx:           *http.Client with retry / ratelimit / circuit breaker / logging
//   - keyedmutex:      per-key mutex, slot GC when idle
//   - lru:             generic LRU cache; optional TTL; sharded variant
//   - pool:            bounded goroutine pool with panic handler & Stats
//   - priorityqueue:   generic min/max heap over container/heap
//   - random:          random strings; SecureString uses crypto/rand
//   - ratelimit:       token bucket and leaky bucket limiters
//   - rotatingwriter:  io.Writer with daily or size-based rotation
//   - shutdown:        signal-driven graceful shutdown coordinator
//   - slogctx:         log/slog companion to zapctx
//   - taskgroup:       errgroup + concurrency limit + panic recovery
//   - ttlmap:          concurrent map with per-entry TTL and OnExpire
//   - zapctx:          *zap.Logger and structured fields via context.Context
//
// See README.md for the package index and examples/ for end-to-end programs
// that combine several packages.
package goost
