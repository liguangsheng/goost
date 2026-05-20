package benchmark

import (
	"testing"

	"github.com/dgraph-io/ristretto/v2"
)

func newRistretto(b *testing.B) *ristretto.Cache[string, any] {
	c, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: int64(size * 10),
		MaxCost:     int64(size),
		BufferItems: 64,
	})
	if err != nil {
		b.Fatal(err)
	}
	return c
}

func Benchmark_ristretto_Set(b *testing.B) {
	c := newRistretto(b)
	defer c.Close()
	for i := range b.N {
		c.Set(key(i), value(i), 1)
	}
}

func Benchmark_ristretto_Get(b *testing.B) {
	c := newRistretto(b)
	defer c.Close()
	for i := range b.N {
		if i%2 == 0 {
			c.Set(key(i), value(i), 1)
		}
	}
	c.Wait()
	b.ResetTimer()
	for i := range b.N {
		c.Get(key(i))
	}
}
