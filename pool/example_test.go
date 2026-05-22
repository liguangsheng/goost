package pool_test

import (
	"fmt"

	"github.com/liguangsheng/goost/pool"
)

func ExampleNewPool() {
	p, err := pool.NewPool(1, 2, 1)
	if err != nil {
		panic(err)
	}

	results := make(chan string, 2)
	if err := p.Schedule(func() { results <- "first" }); err != nil {
		panic(err)
	}
	if err := p.Schedule(func() { results <- "second" }); err != nil {
		panic(err)
	}

	p.Close()
	close(results)

	for result := range results {
		fmt.Println(result)
	}
	stats := p.Stats()
	fmt.Println(stats.Completed)
	fmt.Println(stats.Closed)

	// Output:
	// first
	// second
	// 2
	// true
}
