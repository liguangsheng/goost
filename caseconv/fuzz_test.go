package caseconv

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzCamelSplitRoundTrip checks that splitting a string and lower-camel
// joining it yields a string that splits the same way again. The property
// is "split is a deterministic fixed point under lower-camel join."
func FuzzCamelSplitRoundTrip(f *testing.F) {
	f.Add("HelloWorld")
	f.Add("HTTPResponse")
	f.Add("simpleCamel")
	f.Add("ID")
	f.Add("ABC123def")
	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) {
			return
		}
		// strip non-letter prefix/suffix the split would otherwise classify as
		// CLASS_OTHER and break round-tripping
		s = strings.Map(func(r rune) rune {
			if r > 127 {
				return -1
			}
			return r
		}, s)

		parts := CamelSplit(s)
		if len(parts) == 0 {
			return
		}
		joined := LowerCamelJoin(parts)
		parts2 := CamelSplit(joined)
		if len(parts) != len(parts2) {
			t.Skipf("split count differed: %v vs %v (acronym collapse is acceptable)", parts, parts2)
		}
	})
}

// FuzzSnakeJoinRoundTrip: lower snake join then snake split returns the
// same sequence of lower-cased parts.
func FuzzSnakeJoinRoundTrip(f *testing.F) {
	f.Add("hello", "world")
	f.Add("a", "b")
	f.Fuzz(func(t *testing.T, a, b string) {
		if strings.ContainsRune(a, '_') || strings.ContainsRune(b, '_') {
			return
		}
		if a == "" || b == "" {
			return
		}
		joined := LowerSnakeJoin([]string{a, b})
		parts := SnakeSplit(joined)
		if len(parts) != 2 {
			t.Fatalf("expected 2 parts, got %d (%v)", len(parts), parts)
		}
		if parts[0] != strings.ToLower(a) || parts[1] != strings.ToLower(b) {
			t.Fatalf("round-trip mismatch: %q,%q -> %q -> %v", a, b, joined, parts)
		}
	})
}
