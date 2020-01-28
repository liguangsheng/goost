package lru

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newTestLRU() *lru {
	return New().Build()
}

func Test_LRU(t *testing.T) {
	c := newTestLRU()
	c.Set("hello", "world")
	ret, ok := c.Get("hello")
	assert.Equal(t, true, ok)
	assert.Equal(t, "world", ret)
	c.Remove("hello")
	_, ok = c.Get("hello")
	assert.Equal(t, false, ok)

	for i := 0; i < 100000; i++ {
		c.Set(string(i), string(i))
	}
}

func Test_LRUOverflow(t *testing.T) {
	c := New().Cap(3).Build()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	c.Set("d", 4)
	_, ok := c.Get("a")
	assert.Equal(t, false, ok)
	c.Get("b")
	c.Set("c", 1)
	_, ok = c.Get("c")
	assert.Equal(t, true, ok)
}

func Test_LRUExpire(t *testing.T) {
	c := newTestLRU()
	c.SetWithExpire("hello", "world", time.Now().Add(time.Second))
	time.Sleep(time.Second * 2)
	_, ok := c.Get("hello")
	assert.Equal(t, false, ok)
}

func Test_LRURace(t *testing.T) {
	c := newTestLRU()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
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
