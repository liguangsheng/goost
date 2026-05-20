package batcher

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CoalescesByTimeWindow(t *testing.T) {
	var batches atomic.Int64
	var seenKeys atomic.Int64

	load := func(ctx context.Context, ids []int) (map[int]string, error) {
		batches.Add(1)
		seenKeys.Add(int64(len(ids)))
		out := make(map[int]string, len(ids))
		for _, id := range ids {
			out[id] = "v" + itoa(id)
		}
		return out, nil
	}

	b := New(load).MaxBatch(1000).MaxWait(20 * time.Millisecond).Build()

	const n = 50
	var wg sync.WaitGroup
	results := make([]string, n)
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v, err := b.Load(context.Background(), i)
			assert.NoError(t, err)
			results[i] = v
		}(i)
	}
	wg.Wait()

	assert.EqualValues(t, 1, batches.Load(), "all 50 keys should fit in one batch")
	assert.EqualValues(t, 50, seenKeys.Load())
	for i, v := range results {
		assert.Equal(t, "v"+itoa(i), v)
	}

	s := b.Stats()
	assert.EqualValues(t, 1, s.Batches)
	assert.EqualValues(t, 50, s.Loads)
	assert.EqualValues(t, 49, s.Coalesced, "first key opens batch; remaining 49 join")
	assert.EqualValues(t, 50, s.MaxBatchSize)
}

func Test_FlushesAtMaxBatch(t *testing.T) {
	var batches atomic.Int64
	load := func(ctx context.Context, ids []int) (map[int]string, error) {
		batches.Add(1)
		out := make(map[int]string, len(ids))
		for _, id := range ids {
			out[id] = "v"
		}
		return out, nil
	}

	// MaxWait deliberately huge so only the size trigger can flush.
	b := New(load).MaxBatch(5).MaxWait(time.Hour).Build()

	start := time.Now()
	var wg sync.WaitGroup
	for i := range 5 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := b.Load(context.Background(), i)
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()
	assert.Less(t, time.Since(start), 500*time.Millisecond,
		"size trigger should flush well before MaxWait elapses")
	assert.EqualValues(t, 1, batches.Load())
}

// Trailing partial batch flushes when MaxWait elapses.
func Test_FlushesAtMaxWait(t *testing.T) {
	var batches atomic.Int64
	load := func(ctx context.Context, ids []int) (map[int]string, error) {
		batches.Add(1)
		out := make(map[int]string, len(ids))
		for _, id := range ids {
			out[id] = "v"
		}
		return out, nil
	}

	b := New(load).MaxBatch(1000).MaxWait(15 * time.Millisecond).Build()

	start := time.Now()
	v, err := b.Load(context.Background(), 1)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, "v", v)
	assert.GreaterOrEqual(t, elapsed, 15*time.Millisecond)
	assert.EqualValues(t, 1, batches.Load())
}

func Test_DuplicateKeysInWindowFireOnce(t *testing.T) {
	var seenKeys atomic.Int64
	load := func(ctx context.Context, ids []int) (map[int]string, error) {
		seenKeys.Add(int64(len(ids)))
		out := make(map[int]string, len(ids))
		for _, id := range ids {
			out[id] = "v"
		}
		return out, nil
	}

	b := New(load).MaxBatch(1000).MaxWait(20 * time.Millisecond).Build()

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = b.Load(context.Background(), 7)
		}()
	}
	wg.Wait()

	assert.EqualValues(t, 1, seenKeys.Load(), "duplicate key should be deduped within the window")
}

func Test_LoadFnError_PropagatesToAllCallers(t *testing.T) {
	want := errors.New("oops")
	load := func(ctx context.Context, ids []int) (map[int]string, error) {
		return nil, want
	}
	b := New(load).MaxWait(5 * time.Millisecond).Build()

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := b.Load(context.Background(), i)
			assert.ErrorIs(t, err, want)
		}(i)
	}
	wg.Wait()
}

func Test_LoadFnPanic_BecomesError(t *testing.T) {
	load := func(ctx context.Context, ids []int) (map[int]string, error) {
		panic("boom")
	}
	b := New(load).MaxWait(5 * time.Millisecond).Build()
	_, err := b.Load(context.Background(), 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func Test_MissingKey_ReturnsErrNotFound(t *testing.T) {
	load := func(ctx context.Context, ids []int) (map[int]string, error) {
		return map[int]string{}, nil
	}
	b := New(load).MaxWait(5 * time.Millisecond).Build()
	_, err := b.Load(context.Background(), 1)
	assert.ErrorIs(t, err, ErrNotFound)
}

func Test_ContextCancelReturnsImmediately(t *testing.T) {
	hold := make(chan struct{})
	defer close(hold)
	load := func(ctx context.Context, ids []int) (map[int]string, error) {
		<-hold
		out := make(map[int]string, len(ids))
		for _, id := range ids {
			out[id] = "v"
		}
		return out, nil
	}
	b := New(load).MaxWait(5 * time.Millisecond).Build()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := b.Load(ctx, 1)
	assert.ErrorIs(t, err, context.Canceled)
}

func Test_LoadMany(t *testing.T) {
	load := func(ctx context.Context, ids []int) (map[int]string, error) {
		out := make(map[int]string, len(ids))
		for _, id := range ids {
			if id == 99 {
				continue // simulate missing
			}
			out[id] = "v" + itoa(id)
		}
		return out, nil
	}
	b := New(load).MaxBatch(100).MaxWait(10 * time.Millisecond).Build()

	vals, errs := b.LoadMany(context.Background(), []int{1, 2, 99})
	assert.Equal(t, "v1", vals[1])
	assert.Equal(t, "v2", vals[2])
	assert.ErrorIs(t, errs[99], ErrNotFound)
	_, ok := vals[99]
	assert.False(t, ok)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
