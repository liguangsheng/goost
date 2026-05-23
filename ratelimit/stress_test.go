package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_StressBucketConcurrentWait(t *testing.T) {
	b := NewBucket(10_000, 100)
	ctx := context.Background()

	var ok atomic.Int64
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				if err := b.Wait(ctx, 1); err == nil {
					ok.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	assert.EqualValues(t, 400, ok.Load(), "all waits must succeed with enough rate")
}

func Test_StressLeakyConcurrentWait(t *testing.T) {
	interval := time.Microsecond
	l := NewLeaky(interval)
	ctx := context.Background()

	var ok atomic.Int64
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				if err := l.Wait(ctx); err == nil {
					ok.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	assert.EqualValues(t, 200, ok.Load(), "all waits must succeed with short interval")
}

func Test_StressBucketSnapshotDuringWait(t *testing.T) {
	b := NewBucket(1, 10)
	ctx := context.Background()

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				_ = b.Snapshot()
			}
		}()
	}
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				_ = b.Wait(ctx, 1)
			}
		}()
	}
	wg.Wait()
}
