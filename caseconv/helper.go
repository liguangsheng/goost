package caseconv

import (
	"bytes"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

var acronyms = sync.Map{}

func init() {
	for _, a := range []string{"HTTP", "ID", "WWW", "URL", "DAO", "XML"} {
		acronyms.Store(a, struct{}{})
	}
}

// RegisterAcronym marks s as an acronym so the camel-case routines preserve
// its casing. The comparison is case-insensitive.
func RegisterAcronym(s string) {
	acronyms.Store(strings.ToUpper(s), struct{}{})
}

// UnregisterAcronym removes s from the acronym set.
func UnregisterAcronym(s string) {
	acronyms.Delete(strings.ToUpper(s))
}

// IsAcronym reports whether s (case-insensitive) is registered as an acronym.
func IsAcronym(s string) bool {
	_, ok := acronyms.Load(strings.ToUpper(s))
	return ok
}

// titleFirst returns s with its first rune upper-cased. It is the
// replacement for the deprecated strings.Title used in this package's
// "Title*" join functions.
func titleFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if !unicode.IsLower(r) {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

// SimpleJoin concatenates the entries of a using sep, applying handle to each.
func SimpleJoin(a []string, sep string, handle func(string) string) string {
	if len(a) == 0 {
		return ""
	}
	var buf bytes.Buffer
	buf.WriteString(handle(a[0]))
	for _, s := range a[1:] {
		buf.WriteString(sep)
		buf.WriteString(handle(s))
	}
	return buf.String()
}
