package taskgroup

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_AllSucceed(t *testing.T) {
	g := New(context.Background())
	var sum atomic.Int64
	for i := range 10 {
		g.Go(func(_ context.Context) error {
			sum.Add(int64(i))
			return nil
		})
	}
	assert.NoError(t, g.Wait())
	assert.EqualValues(t, 45, sum.Load())
}

func Test_FirstErrorWins(t *testing.T) {
	g := New(context.Background())
	first := errors.New("first")
	g.Go(func(_ context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return first
	})
	g.Go(func(_ context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return errors.New("second")
	})
	assert.ErrorIs(t, g.Wait(), first)
}

func Test_CauseReportsFirstError(t *testing.T) {
	g := New(context.Background())
	fail := errors.New("fail")

	g.Go(func(_ context.Context) error { return fail })

	assert.ErrorIs(t, g.Wait(), fail)
	assert.ErrorIs(t, g.Cause(), fail)
}

func Test_ErrorCancelsSiblings(t *testing.T) {
	g := New(context.Background())
	fail := errors.New("fail")
	var canceled atomic.Bool

	g.Go(func(_ context.Context) error { return fail })
	g.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			canceled.Store(true)
		case <-time.After(time.Second):
		}
		return nil
	})

	assert.ErrorIs(t, g.Wait(), fail)
	assert.True(t, canceled.Load(), "sibling task should observe cancellation")
}

func Test_WaitCancelsContextOnSuccess(t *testing.T) {
	g := New(context.Background())
	g.Go(func(_ context.Context) error { return nil })

	assert.NoError(t, g.Wait())
	assert.ErrorIs(t, g.Context().Err(), context.Canceled)
}

func Test_GoAfterCancelDoesNotStartTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := New(ctx)

	var ran atomic.Bool
	g.Go(func(_ context.Context) error {
		ran.Store(true)
		return nil
	})

	assert.NoError(t, g.Wait())
	assert.False(t, ran.Load())
}

func Test_Limit(t *testing.T) {
	g := New(context.Background()).WithLimit(2)
	var active, max atomic.Int64

	for range 8 {
		g.Go(func(_ context.Context) error {
			n := active.Add(1)
			for {
				m := max.Load()
				if n <= m || max.CompareAndSwap(m, n) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			active.Add(-1)
			return nil
		})
	}
	assert.NoError(t, g.Wait())
	assert.LessOrEqual(t, max.Load(), int64(2))
}

func Test_PanicBecomesError(t *testing.T) {
	g := New(context.Background())
	g.Go(func(_ context.Context) error { panic("kaboom") })
	err := g.Wait()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic")
}
