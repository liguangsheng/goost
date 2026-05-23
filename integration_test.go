package goost

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/liguangsheng/goost/backoff"
	"github.com/liguangsheng/goost/batcher"
	"github.com/liguangsheng/goost/circuitbreaker"
	"github.com/liguangsheng/goost/fanout"
	"github.com/liguangsheng/goost/httpx"
	"github.com/liguangsheng/goost/lru"
	"github.com/liguangsheng/goost/pool"
	"github.com/liguangsheng/goost/ratelimit"
	"github.com/liguangsheng/goost/shutdown"
	"github.com/liguangsheng/goost/taskgroup"
	"github.com/liguangsheng/goost/ttlmap"
)

func TestBackoffWithCircuitbreaker(t *testing.T) {
	breaker := circuitbreaker.New(circuitbreaker.Config{
		FailureThreshold: 2,
		CooldownPeriod:   50 * time.Millisecond,
	})

	var calls atomic.Int32
	_ = backoff.Retry(context.Background(), &backoff.Backoff{Initial: time.Millisecond}, 3, func(ctx context.Context) error {
		calls.Add(1)
		if breaker.State() == circuitbreaker.StateClosed {
			return fmt.Errorf("fail: %w", circuitbreaker.ErrOpen)
		}
		return nil
	})

	if calls.Load() == 0 {
		t.Error("expected at least one call")
	}
}

func TestPoolWithShutdown(t *testing.T) {
	p, err := pool.NewPool(4, 0, 2)
	if err != nil {
		t.Fatal(err)
	}

	var completed atomic.Int32
	for i := 0; i < 10; i++ {
		_ = p.Schedule(func() { completed.Add(1) })
	}

	mgr := shutdown.NewManager()
	mgr.Add(func() { p.Close() }, shutdown.WithTimeout(5*time.Second))
	mgr.Cleanup()

	if completed.Load() != 10 {
		t.Errorf("expected 10 completed tasks, got %d", completed.Load())
	}
}

func TestRatelimitWithHttpx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	limiter := ratelimit.NewBucket(1, 1)
	client := httpx.New(httpx.Options{Limiter: limiter})

	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
}

func TestLRUWithTTLMapCoexistence(t *testing.T) {
	cache := lru.New[string, int]().Cap(10).Build()
	tm := ttlmap.New[string, int](time.Minute, ttlmap.WithOnExpire(func(k string, v int) {}))
	defer tm.Close()

	cache.Set("a", 1)
	tm.Set("b", 2, 5*time.Minute)

	if v, ok := cache.Get("a"); !ok || v != 1 {
		t.Error("lru cache should have key a")
	}
	if v, ok := tm.Get("b"); !ok || v != 2 {
		t.Error("ttlmap should have key b")
	}
	if _, ok := cache.Get("b"); ok {
		t.Error("lru cache should not have key b")
	}
	if _, ok := tm.Get("a"); ok {
		t.Error("ttlmap should not have key a")
	}
}

func TestBatcherWithFanout(t *testing.T) {
	b := batcher.New(func(ctx context.Context, keys []int) (map[int]string, error) {
		result := make(map[int]string, len(keys))
		for _, k := range keys {
			result[k] = fmt.Sprintf("v%d", k)
		}
		return result, nil
	}).MaxBatch(10).MaxWait(10 * time.Millisecond).Build()

	bc := fanout.New[string]().Buffer(16).Build()
	defer bc.Close()

	sub := bc.Subscribe()

	v, err := b.Load(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	bc.Publish(v)

	select {
	case msg := <-sub.C():
		if msg != "v1" {
			t.Errorf("expected v1, got %s", msg)
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for broadcast")
	}
}

func TestTaskgroupErrorCompat(t *testing.T) {
	g := taskgroup.New(context.Background())
	inner := errors.New("task failed")

	g.Go(func(ctx context.Context) error { return inner })

	err := g.Wait()
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, inner) {
		t.Errorf("expected errors.Is to match inner error, got %v", err)
	}
	if g.Cause() != inner {
		t.Errorf("expected Cause to return inner error, got %v", g.Cause())
	}
}
