package lru

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Keys_MRUOrder(t *testing.T) {
	c := New[string, int]().Cap(4).Build()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	c.Get("a") // a now MRU

	assert.Equal(t, []string{"a", "c", "b"}, c.Keys())
}

func Test_Keys_SkipsExpired(t *testing.T) {
	c := New[string, int]().Cap(4).Build()
	c.Set("live", 1)
	c.SetWithDuration("dead", 2, 10*time.Millisecond)
	time.Sleep(30 * time.Millisecond)

	assert.Equal(t, []string{"live"}, c.Keys())
}

func Test_Range_EarlyStop(t *testing.T) {
	c := New[string, int]().Cap(4).Build()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	var seen []string
	c.Range(func(k string, _ int) bool {
		seen = append(seen, k)
		return len(seen) < 2
	})
	assert.Equal(t, 2, len(seen))
}

func Test_Sharded_RangeAndKeys(t *testing.T) {
	c := New[string, int]().Cap(64).Shards(4, StringHash[string]).BuildSharded()
	for _, k := range []string{"a", "b", "c", "d", "e"} {
		c.Set(k, 1)
	}
	keys := c.Keys()
	sort.Strings(keys)
	assert.Equal(t, []string{"a", "b", "c", "d", "e"}, keys)

	count := 0
	c.Range(func(_ string, _ int) bool {
		count++
		return true
	})
	assert.Equal(t, 5, count)
}
