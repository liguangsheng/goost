package ttlmap

import (
	"fmt"
	"testing"
	"time"
)

func FuzzSetGet(f *testing.F) {
	f.Add(0, "v", 1)
	f.Add(1, "", 0)
	f.Add(-1, "x", 100)

	f.Fuzz(func(t *testing.T, k int, v string, ttl int) {
		m := New[int, string](time.Hour)
		defer m.Close()
		dur := time.Duration(ttl) * time.Millisecond
		if dur < 0 {
			dur = 0
		}
		m.Set(k, v, dur)
		got, ok := m.Get(k)
		if !ok {
			t.Fatalf("Get(%d) not found after Set with TTL %v", k, dur)
		}
		if got != v {
			t.Fatalf("Get(%d) = %q, want %q", k, got, v)
		}
	})
}

func FuzzSetDelete(f *testing.F) {
	f.Add(1)
	f.Add(0)
	f.Add(-42)

	f.Fuzz(func(t *testing.T, k int) {
		m := New[int, int](time.Hour)
		defer m.Close()
		m.Set(k, k, time.Hour)
		m.Delete(k)
		_, ok := m.Get(k)
		if ok {
			t.Fatalf("Get(%d) found after Delete", k)
		}
	})
}

func FuzzSetMany(f *testing.F) {
	f.Add(0)
	f.Fuzz(func(t *testing.T, seed int) {
		m := New[int, string](time.Hour)
		defer m.Close()
		n := (seed % 64) + 1
		if n < 0 {
			n = 1
		}
		for i := range n {
			m.Set(i, fmt.Sprintf("v%d", i), time.Hour)
		}
		for i := range n {
			_, ok := m.Get(i)
			if !ok {
				t.Fatalf("Get(%d) not found after Set(%d)", i, i)
			}
		}
	})
}
