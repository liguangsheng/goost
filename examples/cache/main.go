// cache demonstrates lru + singleflight: even under a thundering herd,
// only one fetch per missing key reaches the slow loader; subsequent
// callers receive the cached value.
//
// Run: go run ./examples/cache
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/liguangsheng/goost/lru"
	"github.com/liguangsheng/goost/singleflight"
)

var loads atomic.Int64

// loadFromOrigin simulates a slow upstream lookup.
func loadFromOrigin(key string) (string, error) {
	loads.Add(1)
	time.Sleep(50 * time.Millisecond)
	return "value-of-" + key, nil
}

func main() {
	cache := lru.New[string, string]().Cap(1024).Build()
	sf := singleflight.NewString[string]()

	get := func(key string) (string, error) {
		if v, ok := cache.Get(key); ok {
			return v, nil
		}
		v, err, _ := sf.Do(key, func() (string, error) {
			return loadFromOrigin(key)
		})
		if err == nil {
			cache.Set(key, v)
		}
		return v, err
	}

	// thundering herd: 100 goroutines hit the same cold key.
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = get("hot-key")
		}()
	}
	wg.Wait()
	fmt.Printf("loads after herd: %d (expected 1)\n", loads.Load())

	// warm cache hit costs almost nothing.
	start := time.Now()
	_, _ = get("hot-key")
	fmt.Printf("warm hit took %s\n", time.Since(start))
}
