# API Conventions

These rules guide new public APIs and cleanup before v1.0.

## Constructors and Zero Values

- Types that own goroutines, timers, locks, maps, queues, or configuration must
  document whether their zero value is usable.
- If the zero value is not useful, prefer an explicit constructor or builder and
  say so in the type doc comment.
- Constructors should reject invalid configuration early when the invalid state
  would otherwise fail later in a goroutine.

## Generics

- Generic APIs should keep type parameters minimal and easy to infer from
  ordinary calls.
- Comparator, loader, callback, and hook signatures should name ownership and
  concurrency expectations in docs when they are not obvious.
- Prefer examples that compile without explicit type arguments when inference is
  expected to work.

## Context, Cancellation, and Lifecycle

- A `context.Context` parameter controls the current operation unless the API
  explicitly documents that it also controls a long-lived object.
- Cancellation should unblock the caller promptly and return `ctx.Err()` or an
  error compatible with `errors.Is` when cancellation is part of the public
  contract.
- Constructors and builders that start goroutines or timers must document the
  matching `Close`, `Stop`, `Wait`, or other release path.
- `Close` and `Stop` methods should be safe to call more than once unless the
  package documents a narrower contract.
- Background goroutines should stop on the documented lifecycle path without
  requiring callers to rely on garbage collection.

## Observability

- `Stats`, `Snapshot`, and similar values should be immutable snapshots from the
  caller's perspective, not live views into internal mutable state.
- Field docs should say whether each value is a current gauge, cumulative
  counter, configuration value, or derived value.
- Snapshot methods must be safe to call concurrently with normal operations when
  the owning type itself is safe for concurrent use.
- Callback and hook APIs should document whether they run synchronously, whether
  they may block progress, and whether panics are recovered.

## Errors and Panics

- Sentinel errors use the `Err` prefix (e.g., `ErrPoolClosed`, `ErrOpen`).
- Sentinel error messages use a package-name prefix (e.g., `"pool: closed"`,
  `"circuitbreaker: open"`) so the source is clear in logs.
- Sentinel errors should be compatible with standard `errors.Is`.
- Wrapped errors should preserve `errors.Is` / `errors.As` behavior via
  `fmt.Errorf("pkg: %w", err)`.
- If a package recovers panics, document that boundary and expose the recovered
  panic as an error or callback value.
- Context cancellation should be returned as a normal error when cancellation is
  part of the public operation.

## Hooks and Callbacks

- Hook and callback functions run synchronously on the caller's goroutine or the
  package's internal goroutine.  Each package must document which.
- Panics inside hooks must not crash the host operation. Packages should recover
  hook panics and keep the host operation's normal result.
- Hooks that may block progress must document the blocking risk so callers can
  decide whether to delegate to a goroutine.
