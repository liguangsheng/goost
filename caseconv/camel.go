package caseconv

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"
)

// runeClass categorizes a rune for word-boundary detection.
type runeClass int

const (
	classLower runeClass = iota
	classUpper
	classDigit
	classOther
)

// prevWordCase tracks the casing produced for the previous word so adjacent
// acronyms don't all get UPPER-cased.
type prevWordCase int

const (
	prevNone prevWordCase = iota
	prevUpper
	prevLower
	prevTitle
)

// CamelSplit splits a camel-case string into its component words.
// Invalid UTF-8 is returned unchanged in a single-element slice.
func CamelSplit(src string) []string {
	if !utf8.ValidString(src) {
		return []string{src}
	}
	if src == "" {
		return []string{}
	}

	var runes [][]rune
	var class, lastClass runeClass
	for i, r := range src {
		switch {
		case unicode.IsLower(r):
			class = classLower
		case unicode.IsUpper(r):
			class = classUpper
		case unicode.IsDigit(r):
			class = classDigit
		default:
			class = classOther
		}
		if i > 0 && class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	// Handle UPPER->lower boundaries, e.g. "PDFL"+"oader" -> "PDF"+"Loader".
	for i := 0; i < len(runes)-1; i++ {
		if IsAcronym(string(runes[i])) {
			continue
		}
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}

	out := make([]string, 0, len(runes))
	for _, s := range runes {
		if len(s) > 0 {
			out = append(out, string(s))
		}
	}
	return out
}

// CamelJoin joins parts into a single camel-case string. When upper is true,
// the result is UpperCamelCase (Pascal case); otherwise lowerCamelCase.
// Registered acronyms are rendered in upper case, except that two adjacent
// acronyms produce one upper + one lower so the boundary stays readable.
func CamelJoin(parts []string, upper bool) string {
	if len(parts) == 0 {
		return ""
	}

	var buf bytes.Buffer
	var prev prevWordCase

	first := parts[0]
	if upper {
		if IsAcronym(first) {
			first = strings.ToUpper(first)
		} else {
			first = titleFirst(strings.ToLower(first))
		}
	} else {
		first = strings.ToLower(first)
	}
	buf.WriteString(first)

	for _, part := range parts[1:] {
		var word string
		if IsAcronym(part) {
			if prev == prevUpper {
				word = strings.ToLower(part)
				prev = prevLower
			} else {
				word = strings.ToUpper(part)
				prev = prevUpper
			}
		} else {
			word = titleFirst(strings.ToLower(part))
			prev = prevTitle
		}
		buf.WriteString(word)
	}
	return buf.String()
}

// UpperCamelJoin joins parts with the first letter of each word capitalized.
func UpperCamelJoin(parts []string) string { return CamelJoin(parts, true) }

// LowerCamelJoin joins parts with the first word in lowercase.
func LowerCamelJoin(parts []string) string { return CamelJoin(parts, false) }

// PascalSplit is an alias of CamelSplit kept for readability.
func PascalSplit(s string) []string { return CamelSplit(s) }

// PascalJoin is an alias of UpperCamelJoin kept for readability.
func PascalJoin(a []string) string { return UpperCamelJoin(a) }
