package keyedmutex

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DifferentKeysRunConcurrently(t *testing.T) {
	m := New[string]()
	start := make(chan struct{})
	var inFlight, peak atomic.Int32

	var wg sync.WaitGroup
	for _, k := range []string{"a", "b", "c", "d"} {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			<-start
			m.Lock(k)
			defer m.Unlock(k)
			cur := inFlight.Add(1)
			for {
				p := peak.Load()
				if cur <= p || peak.CompareAndSwap(p, cur) {
					break
				}
			}
			time.Sleep(30 * time.Millisecond)
			inFlight.Add(-1)
		}(k)
	}
	close(start)
	wg.Wait()
	assert.EqualValues(t, 4, peak.Load(), "different keys should run in parallel")
}

func Test_SameKeySerializes(t *testing.T) {
	m := New[string]()
	var inFlight, peak atomic.Int32
	const n = 8

	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Lock("k")
			defer m.Unlock("k")
			cur := inFlight.Add(1)
			for {
				p := peak.Load()
				if cur <= p || peak.CompareAndSwap(p, cur) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			inFlight.Add(-1)
		}()
	}
	wg.Wait()
	assert.EqualValues(t, 1, peak.Load(), "same key should never overlap")
}

func Test_TryLock(t *testing.T) {
	m := New[string]()
	require.True(t, m.TryLock("k"))
	assert.False(t, m.TryLock("k"), "second TryLock on same key must fail")
	assert.True(t, m.TryLock("other"), "TryLock on different key succeeds")
	m.Unlock("k")
	m.Unlock("other")
	assert.EqualValues(t, 0, m.Len())
}

func Test_TryLockFailDoesNotLeakSlot(t *testing.T) {
	m := New[string]()
	m.Lock("k")
	for range 10 {
		assert.False(t, m.TryLock("k"))
	}
	assert.EqualValues(t, 1, m.Len(), "failed TryLock must not bump refs permanently")
	m.Unlock("k")
	assert.EqualValues(t, 0, m.Len())
}

func Test_LockContextCancel(t *testing.T) {
	m := New[string]()
	m.Lock("k")
	defer m.Unlock("k")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := m.LockContext(ctx, "k")
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.EqualValues(t, 1, m.Len(), "failed LockContext must clean up its ref")
}

func Test_LockContextSucceedsWhenAvailable(t *testing.T) {
	m := New[string]()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, m.LockContext(ctx, "k"))
	m.Unlock("k")
	assert.EqualValues(t, 0, m.Len())
}

func Test_WithLock(t *testing.T) {
	m := New[string]()
	var ran bool
	err := m.WithLock(context.Background(), "k", func() error {
		ran = true
		// Inside fn, holding the lock means TryLock from elsewhere fails.
		assert.False(t, m.TryLock("k"))
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, ran)
	assert.EqualValues(t, 0, m.Len(), "WithLock must release the lock")
}

func Test_WithLockReleasesOnPanic(t *testing.T) {
	m := New[string]()
	assert.Panics(t, func() {
		_ = m.WithLock(context.Background(), "k", func() error {
			panic("boom")
		})
	})
	assert.EqualValues(t, 0, m.Len(), "WithLock must release the lock even on panic")
}

func Test_UnlockUnlockedPanics(t *testing.T) {
	m := New[string]()
	assert.Panics(t, func() { m.Unlock("never-locked") })
}

func Test_DoubleUnlockPanics(t *testing.T) {
	m := New[string]()
	m.Lock("k")
	m.Unlock("k")
	assert.Panics(t, func() { m.Unlock("k") })
}

func Test_HighChurnDoesNotLeak(t *testing.T) {
	m := New[int]()

	// Hammer many short-lived keys; map should drain to empty.
	const workers = 8
	const perWorker = 200
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := range perWorker {
				k := w*perWorker + i
				m.Lock(k)
				m.Unlock(k)
			}
		}(w)
	}
	wg.Wait()
	assert.EqualValues(t, 0, m.Len(), "all slots must be released")
}

func Test_HandoffPreservesSlotWhenContended(t *testing.T) {
	m := New[string]()
	m.Lock("k")
	got := make(chan struct{})
	go func() {
		m.Lock("k")
		m.Unlock("k")
		close(got)
	}()
	// Wait until the waiter has registered.
	require.Eventually(t, func() bool { return m.Len() == 1 },
		time.Second, time.Millisecond)
	m.Unlock("k")
	<-got
	assert.EqualValues(t, 0, m.Len())
}
