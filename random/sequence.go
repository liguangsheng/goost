package random

import (
	"math/bits"
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

var g Sequence

func String(n uint, charsets string) string {
	return g.Next(n, charsets)
}

// Charsets
const (
	Uppercase         = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Lowercase         = "abcdefghijklmnopqrstuvwxyz"
	Alphabetic        = Uppercase + Lowercase
	Numeric           = "0123456789"
	Alphanumeric      = Alphabetic + Numeric
	HumanAlphanumeric = "23456789abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ" // without 1iIlLo0
	Symbols           = "`" + `~!@#$%^&*()-_+={}[]|\;:"<>,./?`
	Hex               = Numeric + "abcdef"
)

type Sequence struct {
	src rand.Source
	mu  sync.Mutex
}

func (s *Sequence) Next(n uint, charsets string) string {
	if s.src == nil {
		s.src = rand.NewSource(time.Now().UnixNano())
	}

	// charsets = repeatToNextPowerOfTwo(charsets)

	bits := uint(bits.Len(uint(len(charsets))))
	mask := int64(1<<bits - 1)
	max := 63 / bits
	b := make([]byte, n)

	for i, cache, remain := int(n-1), s.rand(), max; i >= 0; {
		if remain == 0 {
			cache, remain = s.rand(), max
		}
		if idx := int(cache & mask); idx < len(charsets) {
			b[i] = charsets[idx]
			i--
		}
		cache >>= bits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}

func (s *Sequence) rand() int64 {
	s.mu.Lock()
	r := s.src.Int63()
	s.mu.Unlock()
	return r
}
