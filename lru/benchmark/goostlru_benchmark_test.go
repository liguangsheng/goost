package benchmark

import (
	"strconv"
	"testing"

	golru "github.com/liguangsheng/goost/lru"
)

const size = 1000 * 1000

var values = []any{
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
	return strconv.Itoa(i)
}

func value(i int) any {
	return values[i%len(values)]
}

func Benchmark_goostlru_Set(b *testing.B) {
	c := golru.New[string, any]().Cap(size).Safe(true).Build()
	for i := range b.N {
		c.Set(key(i), value(i))
	}
}

func Benchmark_goostlru_Get(b *testing.B) {
	c := golru.New[string, any]().Cap(size).Safe(true).Build()
	for i := range size {
		if i%2 == 0 {
			c.Set(key(i), value(i))
		}
	}
	b.ResetTimer()
	for i := range b.N {
		v, ok := c.Get(key(i))
		if ok {
			_ = v
		}
	}
}

func Benchmark_goostlru_UnsafeSet(b *testing.B) {
	c := golru.New[string, any]().Cap(size).Safe(false).Build()
	for i := range b.N {
		c.Set(key(i), value(i))
	}
}

func Benchmark_goostlru_UnsafeGet(b *testing.B) {
	c := golru.New[string, any]().Cap(size).Safe(false).Build()
	for i := range size {
		if i%2 == 0 {
			c.Set(key(i), value(i))
		}
	}
	b.ResetTimer()
	for i := range b.N {
		v, ok := c.Get(key(i))
		if ok {
			_ = v
		}
	}
}
