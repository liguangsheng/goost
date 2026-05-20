// Package clock abstracts time-of-day and timers so code that schedules
// work can be tested deterministically.
//
// In production use Real. In tests use Mock: it lets you Advance time
// explicitly and decoupled from wall-clock sleeps.
//
// The Mock.Now method matches the func() time.Time signature accepted
// by the SetClock methods in backoff, ratelimit, ttlmap and
// circuitbreaker, so existing modules can be driven by Mock without API
// changes:
//
//	m := clock.NewMock(time.Unix(0, 0))
//	b := ratelimit.NewBucket(1, 1)
//	b.SetClock(m.Now)
//	m.Advance(time.Second) // refills the bucket
//
// Mock also implements the full Clock interface, including AfterFunc
// and NewTicker, so callback- and ticker-driven code can be tested
// without sleeping.
package clock

import (
	"sync"
	"time"
)

// Clock abstracts the small slice of the time package used across goost.
// The zero interface value is not usable; use Real or NewMock.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
	Sleep(d time.Duration)
	// AfterFunc schedules fn to run after d elapses. fn runs on its own
	// goroutine (matching time.AfterFunc); the returned Timer can be
	// used to cancel before fn fires.
	AfterFunc(d time.Duration, fn func()) Timer
	// NewTicker returns a Ticker that delivers ticks every d. d must be
	// positive. Stop the Ticker to release its resources.
	NewTicker(d time.Duration) Ticker
}

// Timer is a single scheduled callback.
type Timer interface {
	// Stop cancels the callback. Returns true if Stop prevented the
	// callback from running; false if it had already run or Stop had
	// already been called.
	Stop() bool
}

// Ticker periodically delivers a tick on C. Channel C is buffered with
// capacity 1; a tick is dropped if the previous one has not been read,
// matching time.Ticker.
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

// Real returns a Clock backed by the standard library.
func Real() Clock { return realClock{} }

type realClock struct{}

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }
func (realClock) Sleep(d time.Duration)                  { time.Sleep(d) }
func (realClock) AfterFunc(d time.Duration, fn func()) Timer {
	return &realTimer{t: time.AfterFunc(d, fn)}
}
func (realClock) NewTicker(d time.Duration) Ticker {
	return &realTicker{t: time.NewTicker(d)}
}

type realTimer struct{ t *time.Timer }

func (r *realTimer) Stop() bool { return r.t.Stop() }

type realTicker struct{ t *time.Ticker }

func (r *realTicker) C() <-chan time.Time { return r.t.C }
func (r *realTicker) Stop()               { r.t.Stop() }

// Mock is a controllable Clock. It is safe for concurrent use.
//
// Time only moves when Advance or Set is called. Pending events (After
// channels, AfterFunc callbacks, Ticker ticks) fire when their
// deadlines are reached during a Set/Advance call.
type Mock struct {
	mu     sync.Mutex
	now    time.Time
	events []*event
}

type event struct {
	deadline time.Time
	period   time.Duration // 0 for one-shot
	fire     func(now time.Time)
	canceled bool
}

// NewMock returns a Mock starting at start.
func NewMock(start time.Time) *Mock {
	return &Mock{now: start}
}

// Now returns the current mock time.
func (m *Mock) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.now
}

// After returns a channel that fires after d of mock time. If d <= 0
// the channel fires immediately with the current mock time.
func (m *Mock) After(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	m.mu.Lock()
	if d <= 0 {
		ch <- m.now
		m.mu.Unlock()
		return ch
	}
	m.events = append(m.events, &event{
		deadline: m.now.Add(d),
		fire:     func(now time.Time) { ch <- now },
	})
	m.mu.Unlock()
	return ch
}

// Sleep blocks until d of mock time has passed.
func (m *Mock) Sleep(d time.Duration) { <-m.After(d) }

// AfterFunc schedules fn to run after d of mock time has passed. fn
// runs on its own goroutine.
func (m *Mock) AfterFunc(d time.Duration, fn func()) Timer {
	m.mu.Lock()
	e := &event{
		deadline: m.now.Add(d),
		fire:     func(now time.Time) { go fn() },
	}
	m.events = append(m.events, e)
	m.mu.Unlock()
	return &mockTimer{m: m, e: e}
}

// NewTicker returns a Ticker that fires every d of mock time.
func (m *Mock) NewTicker(d time.Duration) Ticker {
	if d <= 0 {
		panic("clock: non-positive interval for NewTicker")
	}
	ch := make(chan time.Time, 1)
	m.mu.Lock()
	e := &event{
		deadline: m.now.Add(d),
		period:   d,
		fire: func(now time.Time) {
			select {
			case ch <- now:
			default:
			}
		},
	}
	m.events = append(m.events, e)
	m.mu.Unlock()
	return &mockTicker{m: m, e: e, ch: ch}
}

// Set jumps the clock to t, firing every event whose deadline <= t in
// chronological order. For periodic events (tickers), the event fires
// once per period boundary it crosses; missed ticks beyond the
// channel's buffer are dropped to match time.Ticker.
func (m *Mock) Set(t time.Time) {
	for {
		m.mu.Lock()
		var (
			pick *event
			idx  int
		)
		for i, e := range m.events {
			if e.canceled {
				continue
			}
			if e.deadline.After(t) {
				continue
			}
			if pick == nil || e.deadline.Before(pick.deadline) {
				pick = e
				idx = i
			}
		}
		if pick == nil {
			m.now = t
			m.mu.Unlock()
			return
		}
		m.now = pick.deadline
		firedAt := pick.deadline
		if pick.period > 0 {
			pick.deadline = pick.deadline.Add(pick.period)
		} else {
			m.events = append(m.events[:idx], m.events[idx+1:]...)
		}
		m.mu.Unlock()
		pick.fire(firedAt)
	}
}

// Advance moves the mock clock forward by d.
func (m *Mock) Advance(d time.Duration) { m.Set(m.Now().Add(d)) }

type mockTimer struct {
	m *Mock
	e *event
}

func (t *mockTimer) Stop() bool {
	t.m.mu.Lock()
	defer t.m.mu.Unlock()
	if t.e.canceled {
		return false
	}
	t.e.canceled = true
	for i, x := range t.m.events {
		if x == t.e {
			t.m.events = append(t.m.events[:i], t.m.events[i+1:]...)
			return true
		}
	}
	return false
}

type mockTicker struct {
	m  *Mock
	e  *event
	ch chan time.Time
}

func (t *mockTicker) C() <-chan time.Time { return t.ch }

func (t *mockTicker) Stop() {
	t.m.mu.Lock()
	defer t.m.mu.Unlock()
	if t.e.canceled {
		return
	}
	t.e.canceled = true
	for i, x := range t.m.events {
		if x == t.e {
			t.m.events = append(t.m.events[:i], t.m.events[i+1:]...)
			return
		}
	}
}
