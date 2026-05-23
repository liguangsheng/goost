package lru

import (
	"fmt"
	"testing"
)

func FuzzSetGet(f *testing.F) {
	f.Add(0, "v")
	f.Add(1, "")
	f.Add(-1, "x")

	f.Fuzz(func(t *testing.T, k int, v string) {
		c := New[int, string]().Cap(100).Build()
		c.Set(k, v)
		got, ok := c.Get(k)
		if !ok {
			t.Fatalf("Get(%d) not found after Set", k)
		}
		if got != v {
			t.Fatalf("Get(%d) = %q, want %q", k, got, v)
		}
	})
}

func FuzzSetEviction(f *testing.F) {
	f.Add(0, 1, 2)
	f.Fuzz(func(t *testing.T, a, b, c int) {
		cache := New[int, int]().Cap(2).Build()
		cache.Set(a, a)
		cache.Set(b, b)
		cache.Set(c, c)
		if cache.Size() > 2 {
			t.Fatalf("size %d exceeds cap 2", cache.Size())
		}
		_, okA := cache.Get(a)
		_, okB := cache.Get(b)
		_, okC := cache.Get(c)
		found := 0
		for _, ok := range []bool{okA, okB, okC} {
			if ok {
				found++
			}
		}
		if found != 2 {
			t.Fatalf("found %d keys, want 2", found)
		}
	})
}

func FuzzPeekNoRecencyUpdate(f *testing.F) {
	f.Add(1, 2, 3)
	f.Fuzz(func(t *testing.T, a, b, c int) {
		if a == b || b == c || a == c {
			return
		}
		cache := New[int, int]().Cap(2).Build()
		cache.Set(a, a)
		cache.Set(b, b)
		cache.Peek(a) // no recency update
		cache.Set(c, c) // evicts a (LRU), not b
		_, okA := cache.Get(a)
		if okA {
			t.Fatal("a should have been evicted")
		}
		_, okB := cache.Get(b)
		if !okB {
			t.Fatal("b should still be present")
		}
	})
}

func FuzzLargeWorkload(f *testing.F) {
	f.Add(0)
	f.Fuzz(func(t *testing.T, seed int) {
		c := New[int, string]().Cap(50).Build()
		for i := range 100 {
			c.Set(i, fmt.Sprintf("v%d", (i+seed)%200))
		}
		if c.Size() > 50 {
			t.Fatalf("size %d exceeds cap 50", c.Size())
		}
	})
}
