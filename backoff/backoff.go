// Package backoff implements exponential backoff with optional jitter and
// a context-aware Retry helper.
package backoff

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

// Backoff produces a sequence of delays following an exponential curve,
// clipped at MaxDelay and optionally jittered.
//
// The zero value yields delays starting at 100 ms, doubling, capped at 30 s.
type Backoff struct {
	// Initial is the first delay. Defaults to 100ms when zero.
	Initial time.Duration
	// Max is the maximum delay. Defaults to 30s when zero.
	Max time.Duration
	// Factor multiplies the delay on each step. Defaults to 2.0.
	Factor float64
	// Jitter is the maximum random fraction added to each delay (0–1).
	// 0.2 means up to ±20%. Defaults to 0 (no jitter).
	Jitter float64
	// Rand, if non-nil, is used to draw jitter in place of math/rand/v2.
	// Useful for deterministic tests; pass a function that returns a value
	// in [0, 1).
	Rand func() float64

	step int
}

// Next returns the next delay and advances the sequence.
func (b *Backoff) Next() time.Duration {
	if b.Initial <= 0 {
		b.Initial = 100 * time.Millisecond
	}
	if b.Max <= 0 {
		b.Max = 30 * time.Second
	}
	if b.Factor <= 0 {
		b.Factor = 2.0
	}

	d := float64(b.Initial)
	for range b.step {
		d *= b.Factor
		if d >= float64(b.Max) {
			d = float64(b.Max)
			break
		}
	}
	b.step++

	if b.Jitter > 0 {
		r := rand.Float64
		if b.Rand != nil {
			r = b.Rand
		}
		// add a random offset in [-Jitter, +Jitter] * d
		j := (r()*2 - 1) * b.Jitter
		d += d * j
	}
	if d < 0 {
		d = 0
	}
	if d > float64(b.Max) {
		d = float64(b.Max)
	}
	return time.Duration(d)
}

// Reset returns the sequence to its initial state.
func (b *Backoff) Reset() { b.step = 0 }

// PermanentError wraps an error so Retry stops immediately and returns it.
type PermanentError struct{ Err error }

func (e *PermanentError) Error() string { return e.Err.Error() }
func (e *PermanentError) Unwrap() error { return e.Err }

// Permanent marks err so Retry returns it without further attempts.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return &PermanentError{Err: err}
}

// Retry calls fn until it returns nil, returns a Permanent error, ctx is
// done, or maxAttempts is reached (0 = unlimited). The Backoff sequence
// dictates the wait between attempts.
func Retry(ctx context.Context, b *Backoff, maxAttempts int, fn func(ctx context.Context) error) error {
	var lastErr error
	for attempt := 1; maxAttempts == 0 || attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return lastErr
			}
			return err
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}
		var perm *PermanentError
		if errors.As(err, &perm) {
			return perm.Err
		}
		lastErr = err

		if maxAttempts > 0 && attempt == maxAttempts {
			break
		}

		t := time.NewTimer(b.Next())
		select {
		case <-ctx.Done():
			t.Stop()
			return lastErr
		case <-t.C:
		}
	}
	return lastErr
}
