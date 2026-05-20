package lru

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ResizeShrinkEvicts(t *testing.T) {
	var evicted []string
	c := New[string, int]().Cap(5).Evict(func(k string, _ int) {
		evicted = append(evicted, k)
	}).Build()
	for _, k := range []string{"a", "b", "c", "d", "e"} {
		c.Set(k, 1)
	}
	c.Resize(2)
	assert.Equal(t, 2, c.Size())
	// LRU-first eviction: a,b,c are oldest among a..e.
	assert.Equal(t, []string{"a", "b", "c"}, evicted)
}

func Test_ResizeGrow(t *testing.T) {
	c := New[string, int]().Cap(2).Build()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Resize(10)
	c.Set("c", 3)
	assert.Equal(t, 3, c.Size())
}

func Test_ResizeUnbounded(t *testing.T) {
	c := New[string, int]().Cap(2).Build()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Resize(0) // unbounded
	for i := range 100 {
		c.Set(string(rune('A'+i%26)), i)
	}
	assert.GreaterOrEqual(t, c.Size(), 2)
}
