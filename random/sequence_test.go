package random

import (
	"sync"
	"testing"
)

func Test_Race(t *testing.T) {
	var s Sequence
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ss := s.Next(uint(i), Uppercase)
			_ = ss
			// t.Log(ss)
		}()
	}
	wg.Wait()
}

func Benchmark_String(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = String(16, HumanAlphanumeric)
	}
}
