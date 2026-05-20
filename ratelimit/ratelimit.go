// Package ratelimit provides two concurrency-safe rate limiters.
//
// Bucket is a token bucket: tokens accrue at a constant rate up to a
// burst capacity. Each call to Allow/Wait consumes one or more tokens.
//
// Leaky is a leaky-bucket queue: requests pace out at a constant rate;
// excess requests are dropped (Allow) or wait (Wait).
//
// Both limiters are zero-dependency and safe for concurrent use.
package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrLimitExceeded is returned by Allow-style APIs when the request would
// exceed the limit and the caller has chosen not to wait.
var ErrLimitExceeded = errors.New("ratelimit: limit exceeded")

// Bucket is a token bucket. The zero value is invalid; use NewBucket.
type Bucket struct {
	mu       sync.Mutex
	rate     float64 // tokens per second
	burst    float64
	tokens   float64
	last     time.Time
	nowFn    func() time.Time
}

// NewBucket returns a token bucket that refills at rate tokens per second
// and holds at most burst tokens. The bucket starts full.
func NewBucket(rate float64, burst int) *Bucket {
	if rate <= 0 {
		rate = 1
	}
	if burst <= 0 {
		burst = 1
	}
	return &Bucket{
		rate:   rate,
		burst:  float64(burst),
		tokens: float64(burst),
		last:   time.Now(),
		nowFn:  time.Now,
	}
}

// SetClock replaces the bucket's clock; useful in tests.
func (b *Bucket) SetClock(fn func() time.Time) {
	b.mu.Lock()
	b.nowFn = fn
	b.last = fn()
	b.mu.Unlock()
}

// Allow reports whether a single token is available right now and, if so,
// consumes it.
func (b *Bucket) Allow() bool { return b.AllowN(1) }

// AllowN consumes n tokens if available; returns false otherwise.
func (b *Bucket) AllowN(n int) bool {
	if n <= 0 {
		return true
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refill()
	if b.tokens >= float64(n) {
		b.tokens -= float64(n)
		return true
	}
	return false
}

// Wait blocks until n tokens are available or ctx is canceled.
func (b *Bucket) Wait(ctx context.Context, n int) error {
	if n <= 0 {
		return nil
	}
	if float64(n) > b.burst {
		return errors.New("ratelimit: n exceeds bucket burst")
	}
	for {
		b.mu.Lock()
		b.refill()
		need := float64(n) - b.tokens
		if need <= 0 {
			b.tokens -= float64(n)
			b.mu.Unlock()
			return nil
		}
		wait := time.Duration(need / b.rate * float64(time.Second))
		b.mu.Unlock()

		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
	}
}

func (b *Bucket) refill() {
	now := b.nowFn()
	elapsed := now.Sub(b.last).Seconds()
	if elapsed <= 0 {
		return
	}
	b.tokens += elapsed * b.rate
	if b.tokens > b.burst {
		b.tokens = b.burst
	}
	b.last = now
}

// Leaky is a leaky-bucket limiter pacing N requests per period.
// The zero value is invalid; use NewLeaky.
type Leaky struct {
	mu       sync.Mutex
	interval time.Duration
	next     time.Time
	nowFn    func() time.Time
}

// NewLeaky returns a leaky bucket that allows at most one request per
// interval (i.e. paces at 1/interval requests per second).
func NewLeaky(interval time.Duration) *Leaky {
	if interval <= 0 {
		interval = time.Second
	}
	return &Leaky{
		interval: interval,
		next:     time.Now(),
		nowFn:    time.Now,
	}
}

// SetClock replaces the bucket's clock; useful in tests.
func (l *Leaky) SetClock(fn func() time.Time) {
	l.mu.Lock()
	l.nowFn = fn
	l.next = fn()
	l.mu.Unlock()
}

// Allow reports whether a request can proceed right now.
func (l *Leaky) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.nowFn()
	if now.Before(l.next) {
		return false
	}
	l.next = now.Add(l.interval)
	return true
}

// Wait blocks until a request is allowed or ctx is canceled.
func (l *Leaky) Wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		now := l.nowFn()
		if !now.Before(l.next) {
			l.next = now.Add(l.interval)
			l.mu.Unlock()
			return nil
		}
		wait := l.next.Sub(now)
		l.mu.Unlock()

		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
	}
}
