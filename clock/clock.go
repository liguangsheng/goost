// Package clock abstracts time-of-day and timers so code that schedules
// work can be tested deterministically.
//
// In production use Real. In tests use Mock: it lets you Advance time
// explicitly and decoupled from wall-clock sleeps.
//
// The Mock.Now method matches the func() time.Time signature accepted by
// the SetClock methods in backoff, ratelimit, ttlmap and circuitbreaker,
// so existing modules can be driven by Mock without API changes:
//
//	m := clock.NewMock(time.Unix(0, 0))
//	b := ratelimit.NewBucket(1, 1)
//	b.SetClock(m.Now)
//	m.Advance(time.Second) // refills the bucket
package clock

import (
	"sort"
	"sync"
	"time"
)

// Clock is the minimum surface used by goost packages.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
	Sleep(d time.Duration)
}

// Real returns a Clock backed by the standard library.
func Real() Clock { return realClock{} }

type realClock struct{}

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }
func (realClock) Sleep(d time.Duration)                  { time.Sleep(d) }

// Mock is a controllable Clock. It is safe for concurrent use.
//
// Time only moves when Advance or Set is called. After and Sleep block
// until enough mock time has passed; Advance unblocks any waiters whose
// deadlines have been reached.
type Mock struct {
	mu      sync.Mutex
	now     time.Time
	waiters []*waiter
}

type waiter struct {
	deadline time.Time
	ch       chan time.Time
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

// After returns a channel that fires after d of mock time. Real wall-clock
// time has no effect; callers must Advance the mock.
func (m *Mock) After(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	m.mu.Lock()
	defer m.mu.Unlock()
	deadline := m.now.Add(d)
	if !deadline.After(m.now) {
		ch <- m.now
		return ch
	}
	m.waiters = append(m.waiters, &waiter{deadline: deadline, ch: ch})
	return ch
}

// Sleep blocks until d of mock time has passed.
func (m *Mock) Sleep(d time.Duration) {
	<-m.After(d)
}

// Set jumps the clock to t, firing every waiter whose deadline <= t.
// Waiters fire in deadline order.
func (m *Mock) Set(t time.Time) {
	m.mu.Lock()
	m.now = t
	due, kept := splitWaiters(m.waiters, t)
	m.waiters = kept
	m.mu.Unlock()

	sort.Slice(due, func(i, j int) bool { return due[i].deadline.Before(due[j].deadline) })
	for _, w := range due {
		w.ch <- w.deadline
	}
}

// Advance moves the mock clock forward by d.
func (m *Mock) Advance(d time.Duration) {
	m.Set(m.Now().Add(d))
}

// splitWaiters returns waiters that fire at or before t and those that remain.
func splitWaiters(in []*waiter, t time.Time) (due, kept []*waiter) {
	for _, w := range in {
		if !w.deadline.After(t) {
			due = append(due, w)
		} else {
			kept = append(kept, w)
		}
	}
	return
}
