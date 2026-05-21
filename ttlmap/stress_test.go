package ttlmap

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_StressSweepOnExpireConcurrentAccess(t *testing.T) {
	var expired atomic.Int64
	m := New[string, int](time.Millisecond, WithOnExpire(func(string, int) {
		expired.Add(1)
	}))
	defer m.Close()

	const workers = 16
	const perWorker = 300
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := range perWorker {
				key := strconv.Itoa(w*perWorker + i)
				m.Set(key, i, time.Millisecond)
				_, _ = m.Get(key)
			}
		}(w)
	}
	wg.Wait()

	assert.Eventually(t, func() bool { return m.Len() == 0 }, 2*time.Second, time.Millisecond)
	assert.Greater(t, expired.Load(), int64(0))
}
