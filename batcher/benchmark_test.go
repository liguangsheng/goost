package batcher

import (
	"context"
	"testing"
	"time"
)

func BenchmarkLoad(b *testing.B) {
	load := func(ctx context.Context, keys []int) (map[int]int, error) {
		out := make(map[int]int, len(keys))
		for _, k := range keys {
			out[k] = k
		}
		return out, nil
	}
	bt := New(load).MaxBatch(64).MaxWait(time.Millisecond).Build()
	bg := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = bt.Load(bg, i)
			i++
		}
	})
}
