package defaultmap

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetConstructs(t *testing.T) {
	var calls atomic.Int64
	m := Make(func(k string) int {
		calls.Add(1)
		return len(k)
	})
	assert.Equal(t, 5, m.Get("hello"))
	assert.Equal(t, 5, m.Get("hello"))
	assert.EqualValues(t, 1, calls.Load())
}

func Test_SetOverrides(t *testing.T) {
	m := Make(func(k string) int { return 0 })
	m.Set("x", 42)
	assert.Equal(t, 42, m.Get("x"))
	assert.True(t, m.Has("x"))
}

func Test_DeleteAndLen(t *testing.T) {
	m := Make(func(k string) int { return 0 })
	m.Set("a", 1)
	m.Set("b", 2)
	assert.Equal(t, 2, m.Len())
	m.Delete("a")
	assert.False(t, m.Has("a"))
	assert.Equal(t, 1, m.Len())
}

func Test_Range(t *testing.T) {
	m := Make(func(k string) int { return 0 })
	m.Set("a", 1)
	m.Set("b", 2)

	seen := map[string]int{}
	m.Range(func(k string, v int) bool {
		seen[k] = v
		return true
	})
	assert.Equal(t, map[string]int{"a": 1, "b": 2}, seen)
}

func Test_RaceConstructorOnce(t *testing.T) {
	var calls atomic.Int64
	m := Make(func(k string) int {
		calls.Add(1)
		return 1
	})
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.Get("k")
		}()
	}
	wg.Wait()
	assert.EqualValues(t, 1, calls.Load())
}
