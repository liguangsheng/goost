package benchmark

import (
	"github.com/bluele/gcache"
	"testing"
)


func Benchmark_gcache_lru_Set(b *testing.B) {
	c := gcache.New(size).LRU().Build()
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(i))
	}
}

func Benchmark_gcache_lru_Get(b *testing.B) {
	c := gcache.New(size).LRU().Build()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			c.Set(key(i), value(i))
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(key(i))
	}
}

func Benchmark_gcache_arc_Set(b *testing.B) {
	c := gcache.New(size).ARC().Build()
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(i))
	}
}

func Benchmark_gcache_arc_Get(b *testing.B) {
	c := gcache.New(size).ARC().Build()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			c.Set(key(i), value(i))
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(key(i))
	}
}
