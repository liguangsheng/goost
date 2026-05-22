package taskgroup_test

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/liguangsheng/goost/taskgroup"
)

func ExampleGroup() {
	g := taskgroup.New(context.Background()).WithLimit(2)
	var sum atomic.Int64

	for _, n := range []int64{1, 2, 3} {
		n := n
		g.Go(func(context.Context) error {
			sum.Add(n)
			return nil
		})
	}

	fmt.Println(g.Wait() == nil)
	fmt.Println(sum.Load())

	// Output:
	// true
	// 6
}

func ExampleResults() {
	g := taskgroup.NewResults[string](context.Background()).WithLimit(1)

	for _, item := range []string{"alpha", "beta"} {
		item := item
		g.Run(func(context.Context) (string, error) {
			return strings.ToUpper(item), nil
		})
	}

	values, err := g.Wait()
	fmt.Println(err == nil)
	fmt.Println(values)

	// Output:
	// true
	// [ALPHA BETA]
}
