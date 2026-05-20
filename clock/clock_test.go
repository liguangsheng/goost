package clock

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RealNow(t *testing.T) {
	c := Real()
	a := c.Now()
	time.Sleep(time.Millisecond)
	b := c.Now()
	assert.True(t, b.After(a))
}

func Test_MockNowAdvance(t *testing.T) {
	start := time.Unix(0, 0)
	m := NewMock(start)
	assert.Equal(t, start, m.Now())
	m.Advance(time.Second)
	assert.Equal(t, start.Add(time.Second), m.Now())
}

func Test_MockAfter(t *testing.T) {
	m := NewMock(time.Unix(0, 0))
	ch := m.After(50 * time.Millisecond)

	select {
	case <-ch:
		t.Fatal("After fired before Advance")
	default:
	}

	m.Advance(100 * time.Millisecond)
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("After did not fire after Advance")
	}
}

func Test_MockAfterMultipleWaiters(t *testing.T) {
	m := NewMock(time.Unix(0, 0))
	ch1 := m.After(10 * time.Millisecond)
	ch2 := m.After(20 * time.Millisecond)
	ch3 := m.After(30 * time.Millisecond)

	m.Advance(25 * time.Millisecond)
	<-ch1
	<-ch2
	select {
	case <-ch3:
		t.Fatal("waiter for 30ms fired at 25ms")
	default:
	}
	m.Advance(10 * time.Millisecond)
	<-ch3
}

func Test_MockSleep(t *testing.T) {
	m := NewMock(time.Unix(0, 0))
	var done sync.WaitGroup
	done.Add(1)
	go func() {
		defer done.Done()
		m.Sleep(100 * time.Millisecond)
	}()
	// Give the goroutine time to register the waiter.
	time.Sleep(10 * time.Millisecond)
	m.Advance(100 * time.Millisecond)
	done.Wait()
}

func Test_MockNowFnAdapter(t *testing.T) {
	m := NewMock(time.Unix(123, 0))
	// Method value is usable as func() time.Time; consume it through a
	// parameter to assert that signature without an inferred-type warning.
	use := func(fn func() time.Time) { assert.Equal(t, time.Unix(123, 0), fn()) }
	use(m.Now)
}

func Test_MockAfterFunc(t *testing.T) {
	m := NewMock(time.Unix(0, 0))
	var fired atomic.Bool
	m.AfterFunc(50*time.Millisecond, func() { fired.Store(true) })

	m.Advance(10 * time.Millisecond)
	assert.False(t, fired.Load(), "fired too early")

	m.Advance(100 * time.Millisecond)
	require.Eventually(t, fired.Load, time.Second, time.Millisecond,
		"AfterFunc callback should fire after deadline")
}

func Test_MockAfterFunc_Stop(t *testing.T) {
	m := NewMock(time.Unix(0, 0))
	var fired atomic.Bool
	timer := m.AfterFunc(50*time.Millisecond, func() { fired.Store(true) })

	assert.True(t, timer.Stop(), "Stop before deadline should return true")
	m.Advance(time.Hour)
	time.Sleep(10 * time.Millisecond)
	assert.False(t, fired.Load(), "stopped timer must not fire")

	assert.False(t, timer.Stop(), "second Stop returns false")
}

func Test_MockAfterFunc_StopAfterFireReturnsFalse(t *testing.T) {
	m := NewMock(time.Unix(0, 0))
	done := make(chan struct{})
	timer := m.AfterFunc(10*time.Millisecond, func() { close(done) })

	m.Advance(20 * time.Millisecond)
	<-done
	// Allow the firing goroutine to finalize state.
	assert.False(t, timer.Stop(), "Stop after fire returns false")
}

func Test_MockTicker(t *testing.T) {
	m := NewMock(time.Unix(0, 0))
	tk := m.NewTicker(10 * time.Millisecond)
	defer tk.Stop()

	// No tick before advancing.
	select {
	case <-tk.C():
		t.Fatal("ticker fired before any time advance")
	default:
	}

	m.Advance(10 * time.Millisecond)
	select {
	case ts := <-tk.C():
		assert.Equal(t, time.Unix(0, 0).Add(10*time.Millisecond), ts)
	case <-time.After(time.Second):
		t.Fatal("first tick did not arrive")
	}

	m.Advance(10 * time.Millisecond)
	select {
	case ts := <-tk.C():
		assert.Equal(t, time.Unix(0, 0).Add(20*time.Millisecond), ts)
	case <-time.After(time.Second):
		t.Fatal("second tick did not arrive")
	}
}

func Test_MockTicker_DropsMissedTicks(t *testing.T) {
	m := NewMock(time.Unix(0, 0))
	tk := m.NewTicker(10 * time.Millisecond)
	defer tk.Stop()

	// Advance enough mock time for 5 ticks without reading. Channel cap
	// is 1, so only one tick should be delivered; the rest drop.
	m.Advance(50 * time.Millisecond)
	<-tk.C()
	select {
	case <-tk.C():
		t.Fatal("missed ticks should be dropped, not buffered")
	case <-time.After(20 * time.Millisecond):
	}
}

func Test_MockTicker_Stop(t *testing.T) {
	m := NewMock(time.Unix(0, 0))
	tk := m.NewTicker(10 * time.Millisecond)
	tk.Stop()
	m.Advance(time.Hour)
	select {
	case <-tk.C():
		t.Fatal("stopped ticker must not fire")
	case <-time.After(20 * time.Millisecond):
	}
	tk.Stop() // double-Stop is a no-op
}

func Test_MockTicker_PanicsOnNonPositive(t *testing.T) {
	m := NewMock(time.Unix(0, 0))
	assert.Panics(t, func() { m.NewTicker(0) })
	assert.Panics(t, func() { m.NewTicker(-time.Second) })
}

func Test_RealAfterFunc(t *testing.T) {
	c := Real()
	done := make(chan struct{})
	c.AfterFunc(time.Millisecond, func() { close(done) })
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("real AfterFunc did not fire")
	}
}

func Test_RealTicker(t *testing.T) {
	c := Real()
	tk := c.NewTicker(time.Millisecond)
	defer tk.Stop()
	select {
	case <-tk.C():
	case <-time.After(time.Second):
		t.Fatal("real ticker did not fire")
	}
}

// fire order is by deadline, not registration order.
func Test_MockEventsFireInDeadlineOrder(t *testing.T) {
	m := NewMock(time.Unix(0, 0))
	var order []int
	var mu sync.Mutex
	record := func(n int) func() {
		return func() {
			mu.Lock()
			order = append(order, n)
			mu.Unlock()
		}
	}
	m.AfterFunc(30*time.Millisecond, record(3))
	m.AfterFunc(10*time.Millisecond, record(1))
	m.AfterFunc(20*time.Millisecond, record(2))

	m.Advance(50 * time.Millisecond)
	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(order) == 3
	}, time.Second, time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []int{1, 2, 3}, order)
}
