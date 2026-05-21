package priorityqueue_test

import (
	"fmt"

	"github.com/liguangsheng/goost/priorityqueue"
)

func Example() {
	q := priorityqueue.New[int](func(a, b int) bool { return a < b })
	q.Push(3)
	q.Push(1)
	q.Push(2)

	for _, v := range q.Drain() {
		fmt.Println(v)
	}

	// Output:
	// 1
	// 2
	// 3
}
