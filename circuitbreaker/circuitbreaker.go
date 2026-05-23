// Package circuitbreaker implements a three-state circuit breaker
// (closed / open / half-open) suitable for protecting downstream calls
// from cascading failures.
package circuitbreaker

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

// State is the current state of the breaker.
type State int32

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	}
	return "unknown"
}

// ErrOpen is returned by Do when the breaker is open and rejects the call.
var ErrOpen = errors.New("circuitbreaker: open")

// Config configures a Breaker.
type Config struct {
	// FailureThreshold consecutive failures (in closed state) trip the
	// breaker. Defaults to 5.
	FailureThreshold int
	// CooldownPeriod is how long the breaker stays open before allowing
	// a single half-open probe. Defaults to 30s.
	CooldownPeriod time.Duration
	// HalfOpenSuccesses is the number of consecutive successes required in
	// the half-open state to close the breaker. Defaults to 1.
	HalfOpenSuccesses int
	// IsFailure decides which errors count as failures. Defaults to
	// "any non-nil error". Return false for, e.g., context.Canceled.
	IsFailure func(error) bool
	// OnStateChange is invoked synchronously whenever the breaker transitions
	// states. Panics are recovered and do not affect the transition.
	OnStateChange func(from, to State)
	// Now overrides the clock; useful for tests.
	Now func() time.Time
}

// Breaker is a state machine that prevents repeated calls to a failing
// downstream when it is unlikely to succeed.
type Breaker struct {
	cfg Config

	state            atomic.Int32 // State
	failures         atomic.Int64
	halfOpenSucc     atomic.Int64
	openedAt         atomic.Int64 // unix nano
	halfOpenInFlight atomic.Bool
}

// Snapshot is a point-in-time view of a breaker.
type Snapshot struct {
	State             State
	Failures          int64
	HalfOpenSuccesses int64
	OpenedAt          time.Time
	CooldownRemaining time.Duration
}

// New constructs a Breaker. Zero-valued config fields fall back to defaults.
func New(cfg Config) *Breaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.CooldownPeriod <= 0 {
		cfg.CooldownPeriod = 30 * time.Second
	}
	if cfg.HalfOpenSuccesses <= 0 {
		cfg.HalfOpenSuccesses = 1
	}
	if cfg.IsFailure == nil {
		cfg.IsFailure = func(err error) bool { return err != nil }
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Breaker{cfg: cfg}
}

// State returns the breaker's current state. It will lazily move from open
// to half-open if the cooldown has elapsed.
func (b *Breaker) State() State {
	s := State(b.state.Load())
	if s == StateOpen {
		if b.cooldownElapsed() {
			b.transition(StateOpen, StateHalfOpen)
			return StateHalfOpen
		}
	}
	return s
}

// Snapshot returns a point-in-time view of the breaker for metrics or logs.
// Like State, it lazily moves from open to half-open if the cooldown elapsed.
func (b *Breaker) Snapshot() Snapshot {
	s := b.State()
	openedAtUnix := b.openedAt.Load()
	snap := Snapshot{
		State:             s,
		Failures:          b.failures.Load(),
		HalfOpenSuccesses: b.halfOpenSucc.Load(),
	}
	if openedAtUnix == 0 || s == StateClosed {
		return snap
	}
	snap.OpenedAt = time.Unix(0, openedAtUnix)
	if s == StateOpen {
		elapsed := b.cfg.Now().Sub(snap.OpenedAt)
		remaining := b.cfg.CooldownPeriod - elapsed
		if remaining > 0 {
			snap.CooldownRemaining = remaining
		}
	}
	return snap
}

// Do executes fn, recording its outcome and updating the breaker. If the
// breaker is open, Do returns ErrOpen without calling fn.
//
// In half-open state, only one probe is permitted at a time; concurrent
// callers see ErrOpen until the probe resolves.
func (b *Breaker) Do(ctx context.Context, fn func(context.Context) error) error {
	switch b.State() {
	case StateOpen:
		return ErrOpen
	case StateHalfOpen:
		if !b.halfOpenInFlight.CompareAndSwap(false, true) {
			return ErrOpen
		}
		defer b.halfOpenInFlight.Store(false)
		return b.run(ctx, fn, true)
	default:
		return b.run(ctx, fn, false)
	}
}

func (b *Breaker) run(ctx context.Context, fn func(context.Context) error, halfOpen bool) error {
	err := fn(ctx)
	if err != nil && b.cfg.IsFailure(err) {
		b.onFailure(halfOpen)
		return err
	}
	b.onSuccess(halfOpen)
	return err
}

func (b *Breaker) onSuccess(halfOpen bool) {
	if halfOpen {
		got := b.halfOpenSucc.Add(1)
		if got >= int64(b.cfg.HalfOpenSuccesses) {
			b.transition(StateHalfOpen, StateClosed)
			b.failures.Store(0)
			b.halfOpenSucc.Store(0)
		}
		return
	}
	b.failures.Store(0)
}

func (b *Breaker) onFailure(halfOpen bool) {
	if halfOpen {
		b.transition(StateHalfOpen, StateOpen)
		b.openedAt.Store(b.cfg.Now().UnixNano())
		b.halfOpenSucc.Store(0)
		return
	}
	n := b.failures.Add(1)
	if n >= int64(b.cfg.FailureThreshold) {
		if b.transition(StateClosed, StateOpen) {
			b.openedAt.Store(b.cfg.Now().UnixNano())
		}
	}
}

// cooldownElapsed must only be called when the breaker is in StateOpen,
// so openedAt has been set by the transition that opened it.
func (b *Breaker) cooldownElapsed() bool {
	openedAt := b.openedAt.Load()
	return b.cfg.Now().UnixNano()-openedAt >= int64(b.cfg.CooldownPeriod)
}

// transition CASes state and fires OnStateChange. Returns whether the
// transition actually happened.
func (b *Breaker) transition(from, to State) bool {
	if !b.state.CompareAndSwap(int32(from), int32(to)) {
		return false
	}
	if b.cfg.OnStateChange != nil {
		safeStateChangeHook(b.cfg.OnStateChange, from, to)
	}
	return true
}

func safeStateChangeHook(fn func(from, to State), from, to State) {
	defer func() {
		_ = recover()
	}()
	fn(from, to)
}
