package keyedmutex

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StressKeyChurnWithContext(t *testing.T) {
	m := New[int]()

	const workers = 32
	const perWorker = 500
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := range perWorker {
				key := w*perWorker + i
				err := m.WithLock(context.Background(), key, func() error { return nil })
				assert.NoError(t, err)
			}
		}(w)
	}
	wg.Wait()
	assert.Equal(t, 0, m.Len())
}
