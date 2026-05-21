package batcher_test

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/liguangsheng/goost/batcher"
)

func Example() {
	load := func(ctx context.Context, keys []int) (map[int]string, error) {
		out := make(map[int]string, len(keys))
		for _, key := range keys {
			out[key] = fmt.Sprintf("user-%d", key)
		}
		return out, nil
	}

	b := batcher.New(load).MaxBatch(3).MaxWait(time.Millisecond).Build()
	vals, errs := b.LoadMany(context.Background(), []int{3, 1, 2})
	if len(errs) != 0 {
		panic(errs)
	}

	keys := make([]int, 0, len(vals))
	for key := range vals {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	for _, key := range keys {
		fmt.Println(vals[key])
	}

	// Output:
	// user-1
	// user-2
	// user-3
}
