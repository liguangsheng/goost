// Package debounce coalesces a burst of Trigger(v) calls into a single
// emit on the output channel: after a Trigger, the Debouncer waits a
// "quiet" duration with no further Trigger before forwarding the most
// recent value.
//
// Typical uses:
//
//   - File watcher: collapse 5 filesystem events in 100ms into one
//     "reload now" signal
//   - UI: act after the user stops typing for 300ms
//   - Config refresh: avoid hammering downstream when the source
//     flaps
//
// debounce differs from a rate limiter: a rate limiter throttles call
// rate, dropping or delaying individual events. debounce defers
// emission until the input stream is quiet, then emits the LATEST
// value. Intermediate Trigger values are discarded.
//
// The Clock is injectable so tests can advance time deterministically.
package debounce

import (
	"sync"
	"time"

	"github.com/liguangsheng/goost/clock"
)

// Debouncer buffers Trigger values and emits the latest one after a
// quiet window.
type Debouncer[T any] struct {
	quiet time.Duration
	clk   clock.Clock

	mu      sync.Mutex
	timer   clock.Timer
	pending T
	has     bool
	gen     int64
	out     chan T
	closed  bool
}

// New returns a Debouncer with the given quiet duration. The default
// output buffer is 1 (latest-wins on slow consumers).
func New[T any](quiet time.Duration) *Debouncer[T] {
	if quiet <= 0 {
		panic("debounce: quiet duration must be > 0")
	}
	return &Debouncer[T]{
		quiet: quiet,
		clk:   clock.Real(),
		out:   make(chan T, 1),
	}
}

// WithClock injects a Clock; useful in tests.
func (d *Debouncer[T]) WithClock(c clock.Clock) *Debouncer[T] {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.clk = c
	return d
}

// Trigger schedules an emit. If quiet elapses with no further Trigger,
// v is sent on C(). A Trigger that lands inside an open window
// replaces the pending value and resets the timer.
//
// After Stop, Trigger is a no-op.
func (d *Debouncer[T]) Trigger(v T) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return
	}
	d.pending = v
	d.has = true
	d.gen++
	g := d.gen
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = d.clk.AfterFunc(d.quiet, func() { d.emit(g) })
}

// C is the receive end. It is closed when Stop is called.
func (d *Debouncer[T]) C() <-chan T { return d.out }

// Stop ends processing. Any pending emit is canceled and the output
// channel is closed. Idempotent.
func (d *Debouncer[T]) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return
	}
	d.closed = true
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
	d.has = false
	close(d.out)
}

func (d *Debouncer[T]) emit(g int64) {
	d.mu.Lock()
	if d.closed || !d.has || g != d.gen {
		// Outdated firing: a fresh Trigger superseded this timer
		// before clean cancellation could prevent it.
		d.mu.Unlock()
		return
	}
	v := d.pending
	d.has = false
	d.timer = nil
	out := d.out
	d.mu.Unlock()

	// Latest-wins on slow consumer: try direct send, fall back to
	// replacing whatever stale value is still buffered.
	select {
	case out <- v:
		return
	default:
	}
	select {
	case <-out:
	default:
	}
	select {
	case out <- v:
	default:
	}
}
