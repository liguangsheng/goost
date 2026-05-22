package random_test

import (
	"fmt"

	"github.com/liguangsheng/goost/random"
)

func ExampleNewSequence() {
	seq := random.NewSequence(func() uint64 { return 0 })

	fmt.Println(seq.Next(8, random.Hex))

	// Output:
	// 00000000
}
