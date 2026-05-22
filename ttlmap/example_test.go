package ttlmap_test

import (
	"fmt"

	"github.com/liguangsheng/goost/ttlmap"
)

func ExampleTTLMap() {
	m := ttlmap.New[string, int](0)
	defer m.Close()

	m.Set("answer", 42, 0)

	value, ok := m.Get("answer")
	fmt.Println(ok)
	fmt.Println(value)
	fmt.Println(m.Len())

	// Output:
	// true
	// 42
	// 1
}
