package itertools_test

import (
	"fmt"

	"github.com/liguangsheng/goost/itertools"
)

func ExampleMap() {
	out := itertools.Map([]int{1, 2, 3}, func(v, _ int) int { return v * 2 })
	fmt.Println(out)
	// Output: [2 4 6]
}

func ExampleFilter() {
	out := itertools.Filter([]int{1, 2, 3, 4}, func(v, _ int) bool { return v%2 == 0 })
	fmt.Println(out)
	// Output: [2 4]
}

func ExampleChunk() {
	fmt.Println(itertools.Chunk([]int{1, 2, 3, 4, 5}, 2))
	// Output: [[1 2] [3 4] [5]]
}
