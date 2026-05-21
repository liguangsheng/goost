package random

import (
	"math/bits"
	"math/rand/v2"
)

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

// String returns a random string of length n using runes from charsets.
// It is safe for concurrent use.
func String(n uint, charsets string) string {
	return defaultSeq.Next(n, charsets)
}

// Sequence draws random strings from a user-supplied uint64 source.
// The zero value is not usable; use NewSequence.
type Sequence struct {
	src func() uint64
}

// NewSequence returns a Sequence backed by source. Pass nil to use the
// concurrency-safe default source from math/rand/v2.
func NewSequence(source func() uint64) *Sequence {
	if source == nil {
		source = rand.Uint64
	}
	return &Sequence{src: source}
}

var defaultSeq = NewSequence(nil)

// Next returns a random string of length n using runes from charsets.
func (s *Sequence) Next(n uint, charsets string) string {
	if n == 0 || len(charsets) == 0 {
		return ""
	}

	letterBits := uint(bits.Len(uint(len(charsets) - 1)))
	if letterBits == 0 {
		letterBits = 1
	}
	mask := uint64(1)<<letterBits - 1
	perDraw := 64 / letterBits

	b := make([]byte, n)
	for i, cache, remain := int(n-1), s.src(), perDraw; i >= 0; {
		if remain == 0 {
			cache, remain = s.src(), perDraw
		}
		if idx := int(cache & mask); idx < len(charsets) {
			b[i] = charsets[idx]
			i--
		}
		cache >>= letterBits
		remain--
	}
	return string(b)
}
