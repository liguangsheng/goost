package pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_NewPoolValidation(t *testing.T) {
	_, err := NewPool(0, 0, 0)
	assert.Error(t, err)
	_, err = NewPool(2, 4, 0)
	assert.Error(t, err)
	_, err = NewPool(2, 0, 4)
	assert.Error(t, err)
}

func Test_Schedule(t *testing.T) {
	p, err := NewPool(4, 0, 0)
	assert.NoError(t, err)
	defer p.Close()

	var n atomic.Int64
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		assert.NoError(t, p.Schedule(func() {
			defer wg.Done()
			n.Add(1)
		}))
	}
	wg.Wait()
	assert.EqualValues(t, 100, n.Load())
}

func Test_ScheduleTimeout(t *testing.T) {
	p, err := NewPool(1, 0, 0)
	assert.NoError(t, err)
	defer p.Close()

	// Saturate the single slot with a long task.
	hold := make(chan struct{})
	assert.NoError(t, p.Schedule(func() { <-hold }))

	err = p.ScheduleTimeout(20*time.Millisecond, func() {})
	assert.True(t, errors.Is(err, ErrScheduleTimeout))

	close(hold)
}

func Test_PanicRecover(t *testing.T) {
	var got atomic.Value
	p, err := NewPool(2, 0, 0, WithPanicHandler(func(r any) {
		got.Store(r)
	}))
	assert.NoError(t, err)
	defer p.Close()

	done := make(chan struct{})
	assert.NoError(t, p.Schedule(func() { panic("boom") }))
	// The pool must remain usable after a panicking task.
	assert.NoError(t, p.Schedule(func() { close(done) }))
	<-done

	// Allow the panic handler to land.
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, "boom", got.Load())
}

func Test_Close(t *testing.T) {
	p, err := NewPool(2, 4, 1)
	assert.NoError(t, err)
	p.Close()

	err = p.Schedule(func() {})
	assert.True(t, errors.Is(err, ErrPoolClosed))
}

func Test_Stats(t *testing.T) {
	p, err := NewPool(2, 4, 1)
	assert.NoError(t, err)
	defer p.Close()

	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		assert.NoError(t, p.Schedule(func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
		}))
	}
	wg.Wait()

	s := p.Stats()
	assert.EqualValues(t, 5, s.Completed)
	assert.EqualValues(t, 0, s.InFlight)
	assert.EqualValues(t, 0, s.Panics)
	assert.GreaterOrEqual(t, s.Workers, 1)
}

func Test_StatsCountsPanics(t *testing.T) {
	p, err := NewPool(2, 0, 0, WithPanicHandler(func(any) {}))
	assert.NoError(t, err)
	defer p.Close()

	var wg sync.WaitGroup
	for range 3 {
		wg.Add(1)
		assert.NoError(t, p.Schedule(func() {
			defer wg.Done()
			panic("boom")
		}))
	}
	wg.Wait()

	s := p.Stats()
	assert.EqualValues(t, 3, s.Panics)
}
