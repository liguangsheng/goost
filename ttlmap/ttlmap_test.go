package ttlmap

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_SetGet(t *testing.T) {
	m := New[string, int](0)
	defer m.Close()
	m.Set("a", 1, time.Minute)
	v, ok := m.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}

func Test_Expires(t *testing.T) {
	m := New[string, int](0)
	defer m.Close()
	m.Set("a", 1, 20*time.Millisecond)
	time.Sleep(40 * time.Millisecond)
	_, ok := m.Get("a")
	assert.False(t, ok)
	assert.Equal(t, 0, m.Len())
}

func Test_NoExpireIfZero(t *testing.T) {
	m := New[string, int](0)
	defer m.Close()
	m.Set("a", 1, 0)
	time.Sleep(20 * time.Millisecond)
	v, ok := m.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}

func Test_BackgroundSweep(t *testing.T) {
	m := New[string, int](20 * time.Millisecond)
	defer m.Close()
	m.Set("a", 1, 10*time.Millisecond)
	m.Set("b", 2, 10*time.Millisecond)
	time.Sleep(80 * time.Millisecond)
	assert.Equal(t, 0, m.Len())
}

func Test_OnExpireOnGet(t *testing.T) {
	var fired []string
	m := New(0, WithOnExpire(func(k string, v int) {
		fired = append(fired, k)
	}))
	defer m.Close()
	m.Set("a", 1, 10*time.Millisecond)
	time.Sleep(30 * time.Millisecond)
	_, ok := m.Get("a")
	assert.False(t, ok)
	assert.Equal(t, []string{"a"}, fired)
}

func Test_OnExpireOnSweep(t *testing.T) {
	var fired []string
	var mu sync.Mutex
	m := New(20*time.Millisecond, WithOnExpire(func(k string, v int) {
		mu.Lock()
		fired = append(fired, k)
		mu.Unlock()
	}))
	defer m.Close()
	m.Set("a", 1, 5*time.Millisecond)
	m.Set("b", 2, 5*time.Millisecond)
	time.Sleep(80 * time.Millisecond)
	mu.Lock()
	got := append([]string{}, fired...)
	mu.Unlock()
	assert.ElementsMatch(t, []string{"a", "b"}, got)
}

func Test_DeleteCloseRace(t *testing.T) {
	m := New[string, int](5 * time.Millisecond)
	var wg sync.WaitGroup
	for w := range 8 {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := range 200 {
				k := strconv.Itoa(w*200 + i)
				m.Set(k, i, 2*time.Millisecond)
				m.Get(k)
				m.Delete(k)
			}
		}(w)
	}
	wg.Wait()
	m.Close()
	// Close is idempotent
	m.Close()
}
