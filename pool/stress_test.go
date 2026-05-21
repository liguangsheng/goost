package pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_StressCloseWhileScheduling(t *testing.T) {
	p, err := NewPool(8, 32, 4)
	assert.NoError(t, err)

	var ran atomic.Int64
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				err := p.ScheduleTimeout(time.Millisecond, func() { ran.Add(1) })
				if errors.Is(err, ErrPoolClosed) {
					return
				}
			}
		}()
	}

	time.Sleep(20 * time.Millisecond)
	p.Close()
	wg.Wait()
	assert.Greater(t, ran.Load(), int64(0))
}
