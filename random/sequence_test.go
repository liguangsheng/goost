package random

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_String(t *testing.T) {
	s := String(16, Alphanumeric)
	assert.Equal(t, 16, len(s))
	for _, r := range s {
		assert.True(t, strings.ContainsRune(Alphanumeric, r))
	}
}

func Test_StringEmpty(t *testing.T) {
	assert.Equal(t, "", String(0, Alphanumeric))
	assert.Equal(t, "", String(8, ""))
}

func Test_StringDistribution(t *testing.T) {
	// Sanity check: drawing many strings from a single-rune charset returns
	// only that rune.
	s := String(100, "A")
	assert.Equal(t, strings.Repeat("A", 100), s)
}

func Test_Race(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 1000 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = String(uint(n%32+1), Uppercase)
		}(i)
	}
	wg.Wait()
}

func Benchmark_String(b *testing.B) {
	for range b.N {
		_ = String(16, HumanAlphanumeric)
	}
}
