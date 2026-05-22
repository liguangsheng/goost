package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_OpensAfterThreshold(t *testing.T) {
	b := New(Config{FailureThreshold: 3, CooldownPeriod: time.Second})
	fail := errors.New("fail")
	for range 3 {
		_ = b.Do(context.Background(), func(_ context.Context) error { return fail })
	}
	assert.Equal(t, StateOpen, b.State())
	assert.ErrorIs(t, b.Do(context.Background(), func(_ context.Context) error { return nil }), ErrOpen)
}

func Test_HalfOpenProbe(t *testing.T) {
	now := time.Unix(0, 0)
	b := New(Config{
		FailureThreshold:  2,
		CooldownPeriod:    10 * time.Millisecond,
		HalfOpenSuccesses: 1,
		Now:               func() time.Time { return now },
	})

	fail := errors.New("fail")
	for range 2 {
		_ = b.Do(context.Background(), func(_ context.Context) error { return fail })
	}
	assert.Equal(t, StateOpen, b.State())

	// Cooldown not yet elapsed.
	assert.ErrorIs(t, b.Do(context.Background(), func(_ context.Context) error { return nil }), ErrOpen)

	now = now.Add(20 * time.Millisecond)
	assert.Equal(t, StateHalfOpen, b.State())

	// Successful probe → closed.
	assert.NoError(t, b.Do(context.Background(), func(_ context.Context) error { return nil }))
	assert.Equal(t, StateClosed, b.State())
}

func Test_SnapshotReportsStateAndCounters(t *testing.T) {
	now := time.Unix(100, 0)
	b := New(Config{
		FailureThreshold:  3,
		CooldownPeriod:    time.Second,
		HalfOpenSuccesses: 2,
		Now:               func() time.Time { return now },
	})

	snap := b.Snapshot()
	assert.Equal(t, StateClosed, snap.State)
	assert.EqualValues(t, 0, snap.Failures)
	assert.EqualValues(t, 0, snap.HalfOpenSuccesses)
	assert.True(t, snap.OpenedAt.IsZero())
	assert.Zero(t, snap.CooldownRemaining)

	fail := errors.New("fail")
	for range 2 {
		_ = b.Do(context.Background(), func(_ context.Context) error { return fail })
	}
	snap = b.Snapshot()
	assert.Equal(t, StateClosed, snap.State)
	assert.EqualValues(t, 2, snap.Failures)
	assert.True(t, snap.OpenedAt.IsZero())

	_ = b.Do(context.Background(), func(_ context.Context) error { return fail })
	snap = b.Snapshot()
	assert.Equal(t, StateOpen, snap.State)
	assert.EqualValues(t, 3, snap.Failures)
	assert.Equal(t, time.Unix(100, 0), snap.OpenedAt)
	assert.Equal(t, time.Second, snap.CooldownRemaining)

	now = now.Add(250 * time.Millisecond)
	snap = b.Snapshot()
	assert.Equal(t, StateOpen, snap.State)
	assert.Equal(t, 750*time.Millisecond, snap.CooldownRemaining)

	now = now.Add(time.Second)
	snap = b.Snapshot()
	assert.Equal(t, StateHalfOpen, snap.State)
	assert.Zero(t, snap.CooldownRemaining)
	assert.Equal(t, time.Unix(100, 0), snap.OpenedAt)

	assert.NoError(t, b.Do(context.Background(), func(_ context.Context) error { return nil }))
	snap = b.Snapshot()
	assert.Equal(t, StateHalfOpen, snap.State)
	assert.EqualValues(t, 1, snap.HalfOpenSuccesses)

	assert.NoError(t, b.Do(context.Background(), func(_ context.Context) error { return nil }))
	snap = b.Snapshot()
	assert.Equal(t, StateClosed, snap.State)
	assert.EqualValues(t, 0, snap.Failures)
	assert.EqualValues(t, 0, snap.HalfOpenSuccesses)
	assert.True(t, snap.OpenedAt.IsZero())
}

func Test_HalfOpenFailureReopens(t *testing.T) {
	now := time.Unix(0, 0)
	b := New(Config{
		FailureThreshold: 1,
		CooldownPeriod:   10 * time.Millisecond,
		Now:              func() time.Time { return now },
	})
	fail := errors.New("fail")
	_ = b.Do(context.Background(), func(_ context.Context) error { return fail })
	now = now.Add(20 * time.Millisecond)
	assert.Equal(t, StateHalfOpen, b.State())

	_ = b.Do(context.Background(), func(_ context.Context) error { return fail })
	assert.Equal(t, StateOpen, b.State())
}

func Test_OnStateChange(t *testing.T) {
	type transition struct{ from, to State }
	var got []transition
	var mu sync.Mutex
	b := New(Config{
		FailureThreshold: 1,
		CooldownPeriod:   time.Millisecond,
		OnStateChange: func(from, to State) {
			mu.Lock()
			got = append(got, transition{from, to})
			mu.Unlock()
		},
	})
	_ = b.Do(context.Background(), func(_ context.Context) error { return errors.New("x") })

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, transition{StateClosed, StateOpen}, got[0])
}

func Test_HalfOpenSingleProbe(t *testing.T) {
	now := time.Unix(0, 0)
	b := New(Config{
		FailureThreshold: 1,
		CooldownPeriod:   time.Millisecond,
		Now:              func() time.Time { return now },
	})
	_ = b.Do(context.Background(), func(_ context.Context) error { return errors.New("x") })
	now = now.Add(time.Second)
	assert.Equal(t, StateHalfOpen, b.State())

	hold := make(chan struct{})
	var inflight atomic.Int64
	go func() {
		_ = b.Do(context.Background(), func(_ context.Context) error {
			inflight.Add(1)
			<-hold
			return nil
		})
	}()

	// Wait for the probe to start.
	for inflight.Load() == 0 {
		time.Sleep(time.Millisecond)
	}

	// A concurrent attempt should be rejected.
	assert.ErrorIs(t,
		b.Do(context.Background(), func(_ context.Context) error { return nil }),
		ErrOpen)

	close(hold)
}

func Test_IsFailureCustom(t *testing.T) {
	b := New(Config{
		FailureThreshold: 2,
		IsFailure: func(err error) bool {
			return !errors.Is(err, context.Canceled)
		},
	})
	for range 2 {
		_ = b.Do(context.Background(), func(_ context.Context) error { return context.Canceled })
	}
	assert.Equal(t, StateClosed, b.State(), "Canceled must not trip the breaker")
}
