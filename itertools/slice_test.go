package itertools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SafeSlice(t *testing.T) {
	a := []int{0, 1, 2, 3, 4}
	assert.Equal(t, []int{1, 2}, SafeSlice(a, 1, 3))
	assert.Equal(t, []int{0, 1, 2, 3, 4}, SafeSlice(a, -1, 100))
	assert.Equal(t, []int{}, SafeSlice(a, 10, 20))
	assert.Equal(t, []int{}, SafeSlice(a, 2, 2))
	var nilSlice []int
	assert.Nil(t, SafeSlice(nilSlice, 0, 1))
}

func Test_Difference(t *testing.T) {
	assert.Equal(t, []int{1, 3}, Difference([]int{1, 2, 3, 4}, []int{2, 4}))
	assert.Equal(t, []int{1, 2}, Difference([]int{1, 2}, nil))
}

func Test_Intersection(t *testing.T) {
	assert.Equal(t, []int{2, 4}, Intersection([]int{1, 2, 3, 4}, []int{2, 4, 5}))
	assert.Equal(t, []int{}, Intersection([]int{1, 2}, nil))
}

func Test_Filter(t *testing.T) {
	out := Filter([]int{1, 2, 3, 4}, func(v, _ int) bool { return v%2 == 0 })
	assert.Equal(t, []int{2, 4}, out)
}

func Test_Map(t *testing.T) {
	out := Map([]int{1, 2, 3}, func(v, _ int) int { return v * 2 })
	assert.Equal(t, []int{2, 4, 6}, out)
}

func Test_Reduce(t *testing.T) {
	sum := Reduce([]int{1, 2, 3, 4}, func(agg, v, _ int) int { return agg + v }, 0)
	assert.Equal(t, 10, sum)
}

func Test_Uniq(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3}, Uniq([]int{1, 2, 2, 3, 1}))
}

func Test_Contains(t *testing.T) {
	assert.True(t, Contains([]int{1, 2, 3}, 2))
	assert.False(t, Contains([]int{1, 2, 3}, 9))
}

func Test_Chunk(t *testing.T) {
	assert.Equal(t, [][]int{{1, 2}, {3, 4}, {5}}, Chunk([]int{1, 2, 3, 4, 5}, 2))
	assert.Nil(t, Chunk([]int{1, 2}, 0))
}
