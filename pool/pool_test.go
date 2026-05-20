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
	for i := 0; i < 100; i++ {
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
	p, err := NewPool(2, 0, 0)
	assert.NoError(t, err)
	defer p.Close()

	done := make(chan struct{})
	assert.NoError(t, p.Schedule(func() { panic("boom") }))
	// The pool must remain usable after a panicking task.
	assert.NoError(t, p.Schedule(func() { close(done) }))
	<-done
}

func Test_Close(t *testing.T) {
	p, err := NewPool(2, 4, 1)
	assert.NoError(t, err)
	p.Close()

	err = p.Schedule(func() {})
	assert.True(t, errors.Is(err, ErrPoolClosed))
}
