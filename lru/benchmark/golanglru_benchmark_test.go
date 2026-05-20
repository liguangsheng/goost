package benchmark

import (
	"testing"

	lru "github.com/hashicorp/golang-lru/v2"
)

func Benchmark_golanglruv2_Set(b *testing.B) {
	c, _ := lru.New[string, any](size)
	for i := range b.N {
		c.Add(key(i), value(i))
	}
}

func Benchmark_golanglruv2_Get(b *testing.B) {
	c, _ := lru.New[string, any](size)
	for i := range b.N {
		if i%2 == 0 {
			c.Add(key(i), value(i))
		}
	}
	b.ResetTimer()
	for i := range b.N {
		c.Get(key(i))
	}
}
