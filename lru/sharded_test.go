package lru

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ShardedRoundTrip(t *testing.T) {
	c := New[string, int]().Cap(64).Shards(4, StringHash[string]).BuildSharded()
	for i := range 100 {
		c.Set(strconv.Itoa(i), i)
	}
	hits := 0
	for i := range 100 {
		if v, ok := c.Get(strconv.Itoa(i)); ok {
			assert.Equal(t, i, v)
			hits++
		}
	}
	// Per-shard cap is 16; 4 shards = ~64 fits, so most should hit. But due
	// to per-shard LRU eviction we just check at least some survive.
	assert.Greater(t, hits, 0)
	assert.LessOrEqual(t, c.Size(), 64)
}

func Test_ShardedRace(t *testing.T) {
	c := New[string, int]().Cap(1024).Shards(8, StringHash[string]).BuildSharded()
	var wg sync.WaitGroup
	for w := range 16 {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := range 1000 {
				k := strconv.Itoa(w*1000 + i)
				c.Set(k, i)
				c.Get(k)
			}
		}(w)
	}
	wg.Wait()
}

func Test_ShardedBuildPanicsWithoutHash(t *testing.T) {
	defer func() {
		assert.NotNil(t, recover())
	}()
	New[string, int]().BuildSharded()
}
