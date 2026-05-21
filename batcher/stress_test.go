package batcher

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_StressCancelAndTimeout(t *testing.T) {
	release := make(chan struct{})
	var calls atomic.Int64
	load := func(ctx context.Context, keys []int) (map[int]int, error) {
		calls.Add(1)
		<-release
		out := make(map[int]int, len(keys))
		for _, key := range keys {
			out[key] = key
		}
		return out, nil
	}
	b := New(load).MaxBatch(64).MaxWait(time.Millisecond).Build()

	const workers = 128
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()
			_, _ = b.Load(ctx, i)
		}(i)
	}
	wg.Wait()
	close(release)

	assert.Eventually(t, func() bool { return calls.Load() > 0 }, time.Second, time.Millisecond)
}
