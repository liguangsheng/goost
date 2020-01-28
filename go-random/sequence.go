package random

import (
	"math/bits"
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

type Sequence = *sequence

type sequence struct {
	letters string
	mask    int64
	src     rand.Source
	bits    uint
	max     uint
	mu      sync.Locker
}

func NewSequence(letters string) *sequence {
	s := &sequence{
		letters: letters,
		src:     rand.NewSource(time.Now().UnixNano()),
	}
	s.bits = uint(bits.Len(uint(len(letters))))
	s.mask = 1<<s.bits - 1
	s.max = 63 / s.bits
	s.mu = &sync.Mutex{}
	return s
}

func (s *sequence) Next(n uint) string {
	b := make([]byte, n)
	cache, remain := s.rand(), s.max
	for i := int(n - 1); i >= 0; {
		if remain == 0 {
			cache, remain = s.rand(), s.max
		}
		if idx := int(cache & s.mask); idx < len(s.letters) {
			b[i] = s.letters[idx]
			i--
		}
		cache >>= s.bits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}

func (s *sequence) rand() int64 {
	s.mu.Lock()
	r := s.src.Int63()
	s.mu.Unlock()
	return r
}
