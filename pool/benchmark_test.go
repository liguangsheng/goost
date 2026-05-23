package pool

import (
	"sync/atomic"
	"testing"
)

func BenchmarkSchedule(b *testing.B) {
	p, _ := NewPool(64, 0, 0)
	defer p.Close()
	var n atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = p.Schedule(func() { n.Add(1) })
		}
	})
}

func BenchmarkScheduleWithQueue(b *testing.B) {
	p, _ := NewPool(4, 256, 4)
	defer p.Close()
	var n atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = p.Schedule(func() { n.Add(1) })
		}
	})
}
