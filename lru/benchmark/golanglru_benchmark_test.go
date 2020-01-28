package benchmark

import (
	"github.com/hashicorp/golang-lru"
	"testing"
)

func Benchmark_golanglru_lru_Set(b *testing.B) {
	c, _ := lru.New(size)
	for i := 0; i < b.N; i++ {
		c.Add(key(i), value(i))
	}
}

func Benchmark_golanglru_lru_Get(b *testing.B) {
	c, _ := lru.New(size)
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			c.Add(key(i), value(i))
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(key(i))
	}
}

func Benchmark_golanglru_arc_Set(b *testing.B) {
	c, _ := lru.NewARC(size)
	for i := 0; i < b.N; i++ {
		c.Add(key(i), value(i))
	}
}

func Benchmark_golanglru_arc_Get(b *testing.B) {
	c, _ := lru.NewARC(size)
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			c.Add(key(i), value(i))
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(key(i))
	}
}