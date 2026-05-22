package lru

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newTestLRU() *Cache[string, string] {
	return New[string, string]().Build()
}

func Test_LRU(t *testing.T) {
	c := newTestLRU()
	c.Set("hello", "world")
	ret, ok := c.Get("hello")
	assert.True(t, ok)
	assert.Equal(t, "world", ret)
	c.Remove("hello")
	_, ok = c.Get("hello")
	assert.False(t, ok)

	for i := range 100000 {
		s := strconv.Itoa(i)
		c.Set(s, s)
	}
}

func Test_LRUOverflow(t *testing.T) {
	c := New[string, int]().Cap(3).Build()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	c.Set("d", 4)
	_, ok := c.Get("a")
	assert.False(t, ok)
	c.Get("b")
	c.Set("c", 1)
	_, ok = c.Get("c")
	assert.True(t, ok)
}

func Test_LRUExpire(t *testing.T) {
	c := newTestLRU()
	c.SetWithExpire("hello", "world", time.Now().Add(time.Millisecond*50))
	time.Sleep(time.Millisecond * 150)
	_, ok := c.Get("hello")
	assert.False(t, ok)
	assert.Equal(t, 0, c.Size(), "expired entry should be evicted on access")
}

func Test_LRUPeek(t *testing.T) {
	c := New[string, int]().Cap(2).Build()
	c.Set("a", 1)
	c.Set("b", 2)
	v, ok := c.Peek("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	c.Set("c", 3)
	_, ok = c.Get("a")
	assert.False(t, ok, "Peek must not refresh recency")
}

func Test_LRUEvictHook(t *testing.T) {
	var evicted []string
	c := New[string, int]().Cap(2).Evict(func(k string, v int) {
		evicted = append(evicted, k)
	}).Build()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	assert.Equal(t, []string{"a"}, evicted)
}

func Test_LRUClear(t *testing.T) {
	c := newTestLRU()
	c.Set("a", "1")
	c.Set("b", "2")
	c.Clear()
	assert.Equal(t, 0, c.Size())
	_, ok := c.Get("a")
	assert.False(t, ok)
}

func Test_LRUSnapshot(t *testing.T) {
	c := New[string, int]().Cap(3).Build()
	c.Set("a", 1)
	c.Set("b", 2)

	assert.Equal(t, Snapshot{Size: 2, Capacity: 3, Shards: 1}, c.Snapshot())

	c.Resize(1)
	assert.Equal(t, Snapshot{Size: 1, Capacity: 1, Shards: 1}, c.Snapshot())
}

func Test_LRUSnapshotUnbounded(t *testing.T) {
	c := New[string, int]().Cap(0).Build()
	c.Set("a", 1)
	c.Set("b", 2)

	assert.Equal(t, Snapshot{Size: 2, Capacity: 0, Shards: 1}, c.Snapshot())
}

func Test_ShardedSnapshot(t *testing.T) {
	c := New[string, int]().Cap(10).Shards(3, StringHash[string]).BuildSharded()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	assert.Equal(t, Snapshot{Size: 3, Capacity: 12, Shards: 4}, c.Snapshot())
}

func Test_LRURace(t *testing.T) {
	c := newTestLRU()
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.Set("hello", "world")
		}()

		go func() {
			defer wg.Done()
			c.Get("hello")
		}()
	}
	wg.Wait()
}
