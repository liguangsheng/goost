package taskgroup

import (
	"context"
	"errors"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_ResultsCollectAll(t *testing.T) {
	g := NewResults[int](context.Background())
	for i := range 5 {
		g.Run(func(_ context.Context) (int, error) {
			return i * i, nil
		})
	}
	values, err := g.Wait()
	assert.NoError(t, err)
	sort.Ints(values)
	assert.Equal(t, []int{0, 1, 4, 9, 16}, values)
}

func Test_ResultsFirstError(t *testing.T) {
	g := NewResults[int](context.Background())
	fail := errors.New("fail")
	g.Run(func(_ context.Context) (int, error) { return 0, fail })
	g.Run(func(_ context.Context) (int, error) { return 1, nil })
	_, err := g.Wait()
	assert.ErrorIs(t, err, fail)
	assert.ErrorIs(t, g.Cause(), fail)
}

func Test_ResultsErrorCancelsSiblings(t *testing.T) {
	g := NewResults[int](context.Background())
	fail := errors.New("fail")
	var canceled atomic.Bool

	g.Run(func(_ context.Context) (int, error) { return 0, fail })
	g.Run(func(ctx context.Context) (int, error) {
		select {
		case <-ctx.Done():
			canceled.Store(true)
		case <-time.After(time.Second):
		}
		return 1, nil
	})

	_, err := g.Wait()
	assert.ErrorIs(t, err, fail)
	assert.True(t, canceled.Load(), "sibling task should observe cancellation")
}

func Test_ResultsLimit(t *testing.T) {
	g := NewResults[int](context.Background()).WithLimit(2)
	var active, max atomic.Int64
	for i := range 10 {
		g.Run(func(_ context.Context) (int, error) {
			n := active.Add(1)
			for {
				m := max.Load()
				if n <= m || max.CompareAndSwap(m, n) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			active.Add(-1)
			return i, nil
		})
	}
	values, err := g.Wait()
	assert.NoError(t, err)
	assert.Equal(t, 10, len(values))
	assert.LessOrEqual(t, max.Load(), int64(2))
}

func Test_ResultsWaitCancelsContextOnSuccess(t *testing.T) {
	g := NewResults[int](context.Background())
	g.Run(func(_ context.Context) (int, error) { return 1, nil })

	values, err := g.Wait()
	assert.NoError(t, err)
	assert.Equal(t, []int{1}, values)
	assert.ErrorIs(t, g.Context().Err(), context.Canceled)
}

func Test_ResultsRunAfterCancelDoesNotStartTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := NewResults[int](ctx)

	var ran atomic.Bool
	g.Run(func(_ context.Context) (int, error) {
		ran.Store(true)
		return 1, nil
	})

	values, err := g.Wait()
	assert.NoError(t, err)
	assert.Empty(t, values)
	assert.False(t, ran.Load())
}

func Test_ResultsPanic(t *testing.T) {
	g := NewResults[int](context.Background())
	g.Run(func(_ context.Context) (int, error) { panic("oops") })
	_, err := g.Wait()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic")
}
