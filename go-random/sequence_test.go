package random

import (
	"sync"
	"testing"
)

func Benchmark_String(b *testing.B) {
	s := NewSequence("23456789abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ")
	for i := 0; i < b.N; i++ {
		s.Next(32)
	}
}

func Benchmark_Number(b *testing.B) {
	s := NewSequence("012345678901234567890123456789")
	for i := 0; i < b.N; i++ {
		s.Next(32)
	}
}

func Test_Race(t *testing.T) {
	s := DefaultSequence()
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Next(uint(i))
		}()
	}
	wg.Wait()
}
