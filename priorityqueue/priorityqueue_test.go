package priorityqueue

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MinHeap_Order(t *testing.T) {
	q := New(func(a, b int) bool { return a < b })
	for _, v := range []int{5, 1, 3, 9, 2, 8, 4} {
		q.Push(v)
	}
	var out []int
	for q.Len() > 0 {
		v, ok := q.Pop()
		assert.True(t, ok)
		out = append(out, v)
	}
	assert.Equal(t, []int{1, 2, 3, 4, 5, 8, 9}, out)
}

func Test_MaxHeap_Order(t *testing.T) {
	q := New(func(a, b int) bool { return a > b })
	for _, v := range []int{5, 1, 3, 9, 2} {
		q.Push(v)
	}
	out := q.Drain()
	assert.Equal(t, []int{9, 5, 3, 2, 1}, out)
}

func Test_Peek(t *testing.T) {
	q := New(func(a, b int) bool { return a < b })
	_, ok := q.Peek()
	assert.False(t, ok, "Peek on empty returns ok=false")

	q.Push(3)
	q.Push(1)
	q.Push(2)
	v, ok := q.Peek()
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	assert.Equal(t, 3, q.Len(), "Peek must not remove")
}

func Test_PopEmpty(t *testing.T) {
	q := New(func(a, b string) bool { return a < b })
	v, ok := q.Pop()
	assert.False(t, ok)
	assert.Equal(t, "", v)
}

func Test_Clear(t *testing.T) {
	q := New(func(a, b int) bool { return a < b })
	for i := range 10 {
		q.Push(i)
	}
	q.Clear()
	assert.Equal(t, 0, q.Len())
	_, ok := q.Pop()
	assert.False(t, ok)

	q.Push(2)
	q.Push(1)
	v, ok := q.Pop()
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}

func Test_CustomTypes(t *testing.T) {
	type job struct {
		name     string
		priority int
	}
	q := New(func(a, b job) bool { return a.priority < b.priority })
	q.Push(job{"a", 3})
	q.Push(job{"b", 1})
	q.Push(job{"c", 2})

	got := q.Drain()
	assert.Equal(t, []string{"b", "c", "a"}, []string{got[0].name, got[1].name, got[2].name})
}

func Test_NilLessPanics(t *testing.T) {
	assert.Panics(t, func() { New[int](nil) })
	assert.Panics(t, func() { NewWithCapacity[int](nil, 8) })
}

func Test_NewWithCapacity(t *testing.T) {
	q := NewWithCapacity(func(a, b int) bool { return a < b }, 64)
	for i := 100; i > 0; i-- {
		q.Push(i)
	}
	// Just verify correctness; capacity itself isn't observable via API.
	out := q.Drain()
	expected := make([]int, 100)
	for i := range expected {
		expected[i] = i + 1
	}
	assert.Equal(t, expected, out)
}

func Test_NewWithCapacityNegative(t *testing.T) {
	q := NewWithCapacity(func(a, b int) bool { return a < b }, -1)
	q.Push(2)
	q.Push(1)

	out := q.Drain()
	assert.Equal(t, []int{1, 2}, out)
}

func Test_DrainEmpty(t *testing.T) {
	q := New(func(a, b int) bool { return a < b })
	out := q.Drain()

	assert.Empty(t, out)
	assert.Equal(t, 0, q.Len())
}

// Stress: random sequence in vs. sorted out matches sort package.
func Test_MatchesSortPackage(t *testing.T) {
	in := []int{42, 17, 8, 99, 4, 23, 56, 11, 88, 1, 73}
	q := New(func(a, b int) bool { return a < b })
	for _, v := range in {
		q.Push(v)
	}
	out := q.Drain()

	expected := append([]int(nil), in...)
	sort.Ints(expected)
	assert.Equal(t, expected, out)
}
