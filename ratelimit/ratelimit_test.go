package ratelimit

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_BucketStartsFull(t *testing.T) {
	b := NewBucket(1, 5)
	for range 5 {
		assert.True(t, b.Allow())
	}
	assert.False(t, b.Allow())
}

func Test_BucketRefill(t *testing.T) {
	now := time.Unix(0, 0)
	b := NewBucket(10, 1)
	b.SetClock(func() time.Time { return now })

	assert.True(t, b.Allow())
	assert.False(t, b.Allow())
	now = now.Add(200 * time.Millisecond) // refills 2 tokens, capped at 1
	assert.True(t, b.Allow())
}

func Test_BucketWait(t *testing.T) {
	b := NewBucket(50, 1)
	assert.True(t, b.Allow())
	start := time.Now()
	assert.NoError(t, b.Wait(context.Background(), 1))
	elapsed := time.Since(start)
	// 1 token / 50 per s = 20ms, allow generous upper bound
	assert.GreaterOrEqual(t, elapsed, 10*time.Millisecond)
	assert.Less(t, elapsed, 200*time.Millisecond)
}

func Test_BucketWaitCancel(t *testing.T) {
	b := NewBucket(1, 1)
	assert.True(t, b.Allow())
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	assert.ErrorIs(t, b.Wait(ctx, 1), context.DeadlineExceeded)
}

func Test_BucketParallel(t *testing.T) {
	// 8 goroutines compete; with a fixed clock and a finite burst, total
	// successes must equal exactly burst.
	now := time.Unix(0, 0)
	b := NewBucket(1, 1000)
	b.SetClock(func() time.Time { return now })

	var ok atomic.Int64
	done := make(chan struct{})
	for range 8 {
		go func() {
			for range 1000 {
				if b.Allow() {
					ok.Add(1)
				}
			}
			done <- struct{}{}
		}()
	}
	for range 8 {
		<-done
	}
	assert.EqualValues(t, 1000, ok.Load())
}

func Test_LeakyAllow(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewLeaky(10 * time.Millisecond)
	l.SetClock(func() time.Time { return now })

	assert.True(t, l.Allow())
	assert.False(t, l.Allow())
	now = now.Add(10 * time.Millisecond)
	assert.True(t, l.Allow())
}

func Test_LeakyWait(t *testing.T) {
	l := NewLeaky(20 * time.Millisecond)
	assert.True(t, l.Allow())
	start := time.Now()
	assert.NoError(t, l.Wait(context.Background()))
	assert.GreaterOrEqual(t, time.Since(start), 10*time.Millisecond)
}

func Test_LeakyWaitCancel(t *testing.T) {
	l := NewLeaky(time.Second)
	assert.True(t, l.Allow())
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	assert.ErrorIs(t, l.Wait(ctx), context.DeadlineExceeded)
}
