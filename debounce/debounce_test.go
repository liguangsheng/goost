package debounce

import (
	"testing"
	"time"

	"github.com/liguangsheng/goost/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_EmitAfterQuiet(t *testing.T) {
	m := clock.NewMock(time.Unix(0, 0))
	d := New[string](100 * time.Millisecond).WithClock(m)
	defer d.Stop()

	d.Trigger("hello")
	select {
	case <-d.C():
		t.Fatal("emit before quiet elapsed")
	default:
	}

	m.Advance(100 * time.Millisecond)
	select {
	case v := <-d.C():
		assert.Equal(t, "hello", v)
	case <-time.After(time.Second):
		t.Fatal("emit did not arrive after quiet elapsed")
	}
}

func Test_BurstCollapsesToLatest(t *testing.T) {
	m := clock.NewMock(time.Unix(0, 0))
	d := New[int](100 * time.Millisecond).WithClock(m)
	defer d.Stop()

	for _, v := range []int{1, 2, 3, 4, 5} {
		d.Trigger(v)
		m.Advance(50 * time.Millisecond) // keep within quiet window
	}

	// Still no emit: every Trigger reset the timer.
	select {
	case <-d.C():
		t.Fatal("debouncer emitted before quiet elapsed")
	default:
	}

	m.Advance(100 * time.Millisecond)
	select {
	case v := <-d.C():
		assert.Equal(t, 5, v, "only the latest value should survive a burst")
	case <-time.After(time.Second):
		t.Fatal("emit did not arrive")
	}
}

func Test_MultipleQuietWindowsEmitEach(t *testing.T) {
	m := clock.NewMock(time.Unix(0, 0))
	d := New[int](100 * time.Millisecond).WithClock(m)
	defer d.Stop()

	d.Trigger(1)
	m.Advance(100 * time.Millisecond)
	require.Eventually(t, func() bool {
		select {
		case v := <-d.C():
			return v == 1
		default:
			return false
		}
	}, time.Second, time.Millisecond)

	d.Trigger(2)
	m.Advance(100 * time.Millisecond)
	require.Eventually(t, func() bool {
		select {
		case v := <-d.C():
			return v == 2
		default:
			return false
		}
	}, time.Second, time.Millisecond)
}

func Test_StopClosesChannel(t *testing.T) {
	d := New[int](100 * time.Millisecond)
	d.Stop()
	_, ok := <-d.C()
	assert.False(t, ok)
}

func Test_StopCancelsPending(t *testing.T) {
	m := clock.NewMock(time.Unix(0, 0))
	d := New[int](100 * time.Millisecond).WithClock(m)

	d.Trigger(42)
	d.Stop()
	m.Advance(time.Hour)

	// C is closed; reading returns zero value and ok=false.
	v, ok := <-d.C()
	assert.False(t, ok)
	assert.Equal(t, 0, v)
}

func Test_TriggerAfterStopIsNoop(t *testing.T) {
	d := New[int](100 * time.Millisecond)
	d.Stop()
	assert.NotPanics(t, func() { d.Trigger(1) })
}

func Test_StopIdempotent(t *testing.T) {
	d := New[int](100 * time.Millisecond)
	d.Stop()
	assert.NotPanics(t, d.Stop)
}

func Test_LatestWinsOnSlowConsumer(t *testing.T) {
	m := clock.NewMock(time.Unix(0, 0))
	d := New[int](10 * time.Millisecond).WithClock(m)
	defer d.Stop()

	// Trigger #1, fire, but don't read yet.
	d.Trigger(1)
	m.Advance(10 * time.Millisecond)
	// emit goroutine has to run; give it a moment.
	require.Eventually(t, func() bool { return len(d.C()) == 1 },
		time.Second, time.Millisecond)

	// Trigger #2 + fire while #1 is still buffered. Should replace.
	d.Trigger(2)
	m.Advance(10 * time.Millisecond)
	require.Eventually(t, func() bool {
		// peek by trying to read; latest wins -> 2
		select {
		case v := <-d.C():
			return v == 2
		default:
			return false
		}
	}, time.Second, time.Millisecond)
}

func Test_NewPanicsOnNonPositiveQuiet(t *testing.T) {
	assert.Panics(t, func() { New[int](0) })
	assert.Panics(t, func() { New[int](-time.Second) })
}

func Test_RealClockE2E(t *testing.T) {
	d := New[string](10 * time.Millisecond)
	defer d.Stop()

	d.Trigger("a")
	d.Trigger("b")
	d.Trigger("c")

	select {
	case v := <-d.C():
		assert.Equal(t, "c", v, "burst collapses to latest under real clock too")
	case <-time.After(time.Second):
		t.Fatal("real-clock debouncer did not emit")
	}
}
