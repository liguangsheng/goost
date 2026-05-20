package clock

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
