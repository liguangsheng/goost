package benchmark

import (
	"testing"

	"github.com/bluele/gcache"
)

func Benchmark_gcache_lru_Set(b *testing.B) {
	c := gcache.New(size).LRU().Build()
	for i := range b.N {
		_ = c.Set(key(i), value(i))
	}
}

func Benchmark_gcache_lru_Get(b *testing.B) {
	c := gcache.New(size).LRU().Build()
	for i := range b.N {
		if i%2 == 0 {
			_ = c.Set(key(i), value(i))
		}
	}
	b.ResetTimer()
	for i := range b.N {
		_, _ = c.Get(key(i))
	}
}

func Benchmark_gcache_arc_Set(b *testing.B) {
	c := gcache.New(size).ARC().Build()
	for i := range b.N {
		_ = c.Set(key(i), value(i))
	}
}

func Benchmark_gcache_arc_Get(b *testing.B) {
	c := gcache.New(size).ARC().Build()
	for i := range b.N {
		if i%2 == 0 {
			_ = c.Set(key(i), value(i))
		}
	}
	b.ResetTimer()
	for i := range b.N {
		_, _ = c.Get(key(i))
	}
}
