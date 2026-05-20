# backoff

Exponential backoff with optional jitter and a context-aware `Retry` helper.

```go
b := &backoff.Backoff{
    Initial: 200 * time.Millisecond,
    Max:     10 * time.Second,
    Factor:  2.0,
    Jitter:  0.2, // ±20% randomness
}

err := backoff.Retry(ctx, b, 5, func(ctx context.Context) error {
    return doSomething(ctx)
})
```

Return `backoff.Permanent(err)` from the callback to abort retries
immediately. Pass `maxAttempts = 0` for unlimited retries (governed by
`ctx`).

`Backoff.Rand` accepts a function returning `[0, 1)` so tests can fix the
jitter. See also [`clock`](../clock) for advancing time deterministically
in code that consumes the `Backoff.Next()` durations.
