// cache demonstrates lru + x/sync/singleflight: even under a thundering
// herd, only one fetch per missing key reaches the slow loader;
// subsequent callers receive the cached value.
//
// Run from examples/: go run ./cache
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/liguangsheng/goost/lru"
	"golang.org/x/sync/singleflight"
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
	var sf singleflight.Group

	get := func(key string) (string, error) {
		if v, ok := cache.Get(key); ok {
			return v, nil
		}
		raw, err, _ := sf.Do(key, func() (any, error) {
			return loadFromOrigin(key)
		})
		if err != nil {
			return "", err
		}
		v := raw.(string)
		cache.Set(key, v)
		return v, nil
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

	if v, _ := get("hot-key"); v == "value-of-hot-key" {
		fmt.Println("warm hit reused cached value")
	}
}
