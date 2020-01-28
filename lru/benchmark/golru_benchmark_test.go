package benchmark

import (
	golru "github.com/liguangsheng/go-lru"
	"testing"
)

const size = 1000 * 1000

var values = []interface{}{
	"0123456789",
	"0123456789abcdefghijklmnopqrstuvwxyz",
	`longtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtext
longtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtext
longtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtext
longtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtext
longtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtext
longtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtext
longtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtext
longtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtext
longtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtext
longtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtextlongtext`,
	0, 1, 2, 3,
	3.1, 3.14, 3.141, 3.1415,
}

func key(i int) string {
	return string(i)
}

func value(i int) interface{} {
	return values[i%len(values)]
}

func Benchmark_golru_Set(b *testing.B) {
	c := golru.New().Cap(size).Safe(true).Build()
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(i))
	}

}

func Benchmark_golru_Get(b *testing.B) {
	c := golru.New().Cap(size).Safe(true).Build()
	for i := 0; i < size; i++ {
		if i%2 == 0 {
			c.Set(key(i), value(i))
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value, ok := c.Get(string(i))
		if ok {
			_ = value
		}
	}
}

func Benchmark_golru_UnsafeSet(b *testing.B) {
	c := golru.New().Cap(size).Safe(false).Build()
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(i))
	}

}

func Benchmark_golru_UnsafeGet(b *testing.B) {
	c := golru.New().Cap(size).Safe(false).Build()
	for i := 0; i < size; i++ {
		if i%2 == 0 {
			c.Set(key(i), value(i))
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value, ok := c.Get(string(i))
		if ok {
			_ = value
		}
	}
}
