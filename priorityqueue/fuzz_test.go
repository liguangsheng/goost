package priorityqueue

import (
	"math/rand"
	"testing"
)

func FuzzPushPop(f *testing.F) {
	f.Add(0, 0)
	f.Add(1, -1)
	f.Add(100, 50)
	f.Add(-100, -200)

	f.Fuzz(func(t *testing.T, a, b int) {
		pq := New(func(x, y int) bool { return x < y })
		pq.Push(a)
		pq.Push(b)
		if pq.Len() != 2 {
			t.Fatalf("len = %d, want 2", pq.Len())
		}
		v, ok := pq.Pop()
		if !ok {
			t.Fatal("Pop returned false")
		}
		if a <= b && v != a {
			t.Fatalf("first pop = %d, want %d", v, a)
		}
		if b < a && v != b {
			t.Fatalf("first pop = %d, want %d", v, b)
		}
	})
}

func FuzzPushMany(f *testing.F) {
	seed := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5, 0}
	for _, v := range seed {
		f.Add(v)
	}

	f.Fuzz(func(t *testing.T, v int) {
		pq := New(func(x, y int) bool { return x < y })
		pq.Push(v)
		pq.Push(v + 1)
		pq.Push(v - 1)

		prev, _ := pq.Pop()
		for pq.Len() > 0 {
			cur, ok := pq.Pop()
			if !ok {
				t.Fatal("unexpected empty")
			}
			if cur < prev {
				t.Fatalf("out of order: %d < %d", cur, prev)
			}
			prev = cur
		}
	})
}

func FuzzDrain(f *testing.F) {
	f.Fuzz(func(t *testing.T, size byte) {
		n := int(size) % 64
		vals := make([]int, n)
		for i := range vals {
			vals[i] = rand.Int()
		}
		pq := New(func(x, y int) bool { return x < y })
		for _, v := range vals {
			pq.Push(v)
		}
		count := 0
		for range pq.Drain() {
			count++
		}
		if count != n {
			t.Fatalf("drained %d, pushed %d", count, n)
		}
	})
}
