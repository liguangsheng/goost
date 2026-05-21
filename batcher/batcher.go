// Package batcher coalesces concurrent per-key requests into a single
// batch call. This is the DataLoader pattern: when many goroutines each
// want one key's worth of data and the underlying store can serve N
// keys in one round-trip much more cheaply than N separate ones,
// Batcher collects keys until a size or time window closes and issues
// one loadFn call.
//
// Batcher solves a different problem than singleflight:
//
//   - singleflight dedupes concurrent calls for the SAME key
//   - batcher coalesces concurrent calls for DIFFERENT keys
//
// They compose: a Batcher already dedupes duplicate keys arriving in
// the same window, so wrapping it in a singleflight is unnecessary.
//
// Example:
//
//	b := batcher.New(loadUsers).
//	    MaxBatch(100).
//	    MaxWait(5 * time.Millisecond).
//	    Build()
//	u, err := b.Load(ctx, 42)
//
//	func loadUsers(ctx context.Context, ids []int) (map[int]*User, error) { ... }
package batcher

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// LoadFunc fetches values for a batch of keys. It must return a map
// with one entry per key it could resolve; missing keys are surfaced to
// callers as ErrNotFound. An error returned here propagates to every
// caller in the batch.
type LoadFunc[K comparable, V any] func(ctx context.Context, keys []K) (map[K]V, error)

// ErrNotFound is returned by Load when the batch resolved but did not
// include an entry for the requested key.
var ErrNotFound = errors.New("batcher: key not found in batch result")

// Stats is a snapshot of Batcher counters.
type Stats struct {
	// Batches is the total number of loadFn invocations.
	Batches int64
	// Loads is the total number of Load calls.
	Loads int64
	// Coalesced is the number of Load calls that joined an existing
	// in-flight batch (i.e. did not start a new one).
	Coalesced int64
	// MaxBatchSize is the largest batch loadFn has been called with.
	MaxBatchSize int64
}

// Builder configures a Batcher. Use New to obtain one.
type Builder[K comparable, V any] struct {
	loadFn   LoadFunc[K, V]
	maxBatch int
	maxWait  time.Duration
	ctx      context.Context
}

// New starts a Builder around fn. Default MaxBatch is 128 and default
// MaxWait is 4ms.
func New[K comparable, V any](fn LoadFunc[K, V]) *Builder[K, V] {
	if fn == nil {
		panic("batcher: load func must not be nil")
	}
	return &Builder[K, V]{
		loadFn:   fn,
		maxBatch: 128,
		maxWait:  4 * time.Millisecond,
		ctx:      context.Background(),
	}
}

// MaxBatch caps the number of keys per loadFn invocation. When a batch
// reaches this size it flushes immediately, before MaxWait elapses.
func (b *Builder[K, V]) MaxBatch(n int) *Builder[K, V] {
	if n > 0 {
		b.maxBatch = n
	}
	return b
}

// MaxWait is the longest time the first key in a batch waits before
// the batch flushes. Lower values reduce latency at the cost of
// smaller batches.
func (b *Builder[K, V]) MaxWait(d time.Duration) *Builder[K, V] {
	if d > 0 {
		b.maxWait = d
	}
	return b
}

// Context provides the context passed to loadFn. It does NOT affect
// individual Load(ctx, ...) calls — each Load uses its own ctx for
// waiting. Default is context.Background().
func (b *Builder[K, V]) Context(ctx context.Context) *Builder[K, V] {
	if ctx != nil {
		b.ctx = ctx
	}
	return b
}

// Build returns a ready Batcher.
func (b *Builder[K, V]) Build() *Batcher[K, V] {
	if b.loadFn == nil {
		panic("batcher: load func must not be nil")
	}
	return &Batcher[K, V]{
		loadFn:   b.loadFn,
		maxBatch: b.maxBatch,
		maxWait:  b.maxWait,
		ctx:      b.ctx,
	}
}

// Batcher coalesces Load(key) calls into batch loadFn calls.
type Batcher[K comparable, V any] struct {
	loadFn   LoadFunc[K, V]
	maxBatch int
	maxWait  time.Duration
	ctx      context.Context

	mu  sync.Mutex
	cur *batch[K, V]

	batches      atomic.Int64
	loads        atomic.Int64
	coalesced    atomic.Int64
	maxBatchSize atomic.Int64
}

type batch[K comparable, V any] struct {
	keys   []K
	set    map[K]struct{}
	done   chan struct{}
	result map[K]V
	err    error
	timer  *time.Timer
}

// Load requests value for key. If a batch window is open, key joins
// that batch; otherwise key starts a new window. Load returns when the
// batch resolves or when ctx is canceled.
//
// If the batch's loadFn returns an error, every Load in the batch
// receives that error. If loadFn succeeds but does not include key in
// its result map, Load returns the zero value and ErrNotFound.
func (b *Batcher[K, V]) Load(ctx context.Context, key K) (V, error) {
	b.loads.Add(1)
	cur, flush := b.enqueue(key)

	if flush {
		go b.run(cur)
	}

	var zero V
	select {
	case <-cur.done:
	case <-ctx.Done():
		return zero, ctx.Err()
	}
	if cur.err != nil {
		return zero, cur.err
	}
	v, ok := cur.result[key]
	if !ok {
		return zero, ErrNotFound
	}
	return v, nil
}

// LoadMany is sugar for issuing multiple Loads concurrently and
// collecting results. It returns whatever it could resolve; per-key
// errors (including ErrNotFound) are returned in errs keyed by the
// same K. If ctx is canceled, LoadMany returns whatever has resolved
// so far and ctx.Err() in errs for the unresolved keys.
func (b *Batcher[K, V]) LoadMany(ctx context.Context, keys []K) (vals map[K]V, errs map[K]error) {
	vals = make(map[K]V, len(keys))
	errs = make(map[K]error)
	if len(keys) == 0 {
		return
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, k := range keys {
		wg.Add(1)
		go func(k K) {
			defer wg.Done()
			v, err := b.Load(ctx, k)
			mu.Lock()
			if err != nil {
				errs[k] = err
			} else {
				vals[k] = v
			}
			mu.Unlock()
		}(k)
	}
	wg.Wait()
	return
}

// Stats returns a counter snapshot.
func (b *Batcher[K, V]) Stats() Stats {
	return Stats{
		Batches:      b.batches.Load(),
		Loads:        b.loads.Load(),
		Coalesced:    b.coalesced.Load(),
		MaxBatchSize: b.maxBatchSize.Load(),
	}
}

// enqueue adds key to the current open batch (or starts a new one) and
// returns that batch along with flush=true if the caller should
// immediately invoke run (e.g. MaxBatch reached). When flush is true,
// the batch's timer has been stopped and b.cur has been cleared.
func (b *Batcher[K, V]) enqueue(key K) (*batch[K, V], bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cur == nil {
		bt := &batch[K, V]{
			set:  make(map[K]struct{}),
			done: make(chan struct{}),
		}
		bt.keys = append(bt.keys, key)
		bt.set[key] = struct{}{}
		b.cur = bt
		bt.timer = time.AfterFunc(b.maxWait, func() {
			b.mu.Lock()
			cur := b.cur
			if cur == bt {
				b.cur = nil
			}
			b.mu.Unlock()
			if cur == bt {
				b.run(bt)
			}
		})
		return bt, false
	}

	cur := b.cur
	if _, dup := cur.set[key]; dup {
		b.coalesced.Add(1)
		return cur, false
	}
	cur.set[key] = struct{}{}
	cur.keys = append(cur.keys, key)
	b.coalesced.Add(1)
	if len(cur.keys) >= b.maxBatch {
		if cur.timer != nil {
			cur.timer.Stop()
		}
		b.cur = nil
		return cur, true
	}
	return cur, false
}

// run executes loadFn for bt's keys and signals waiters. Safe to call
// from either the timer goroutine or the size-trigger path; mutually
// exclusive caller guard is the b.cur==bt check in enqueue.
func (b *Batcher[K, V]) run(bt *batch[K, V]) {
	b.batches.Add(1)
	if n := int64(len(bt.keys)); n > b.maxBatchSize.Load() {
		b.maxBatchSize.Store(n)
	}
	defer func() {
		if r := recover(); r != nil {
			bt.err = fmt.Errorf("batcher: panic in loadFn: %v\n%s", r, debug.Stack())
		}
		close(bt.done)
	}()
	bt.result, bt.err = b.loadFn(b.ctx, bt.keys)
}
