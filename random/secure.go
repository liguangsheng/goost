package random

import (
	"crypto/rand"
	"encoding/binary"
	"math/bits"
)

// SecureString returns a random string of length n drawn from charsets,
// using crypto/rand as its source. Suitable for tokens, salts, and other
// security-sensitive uses. Panics if crypto/rand fails.
func SecureString(n uint, charsets string) string {
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
	for i, cache, remain := int(n-1), secureUint64(), perDraw; i >= 0; {
		if remain == 0 {
			cache, remain = secureUint64(), perDraw
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

// secureUint64 returns 8 cryptographically random bytes as a uint64.
// Panics if the system entropy source fails.
func secureUint64() uint64 {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic("random: crypto/rand failed: " + err.Error())
	}
	return binary.LittleEndian.Uint64(buf[:])
}
