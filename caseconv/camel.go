package caseconv

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	CLASS_LOWER = iota
	CLASS_UPPER
	CLASS_DIGIT
	CLASS_OTHER

	UPPER_PREV
	LOWER_PREV
	TITLE_PREV
)

func CamelSplit(src string) (entries []string) {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	// split into fields based on class of unicode character
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = CLASS_LOWER
		case unicode.IsUpper(r):
			class = CLASS_UPPER
		case unicode.IsDigit(r):
			class = CLASS_DIGIT
		default:
			class = CLASS_OTHER
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if IsAcronym(string(runes[i])) {
			continue
		}
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}
	return
}

func CamelJoin(parts []string, upper bool) string {
	if len(parts) == 0 {
		return ""
	}

	var buffer bytes.Buffer
	var prevWordStatus int

	first := parts[0]
	if upper {
		if IsAcronym(first) {
			first = strings.ToUpper(first)
		} else {
			first = strings.Title(first)
		}
	} else {
		first = strings.ToLower(first)
	}
	buffer.WriteString(first)

	for _, part := range parts[1:] {
		var word string
		if IsAcronym(part) {
			if prevWordStatus == UPPER_PREV {
				word = strings.ToLower(part)
				prevWordStatus = LOWER_PREV
			} else {
				word = strings.ToUpper(part)
				prevWordStatus = UPPER_PREV
			}
		} else {
			word = strings.Title(part)
			prevWordStatus = TITLE_PREV
		}
		buffer.WriteString(word)
	}
	return buffer.String()
}

func UpperCamelJoin(parts []string) string {
	return CamelJoin(parts, true)
}

func LowerCamelJoin(parts []string) string {
	return CamelJoin(parts, false)
}

// Pascal style == Upper Camel Style
func PascalSplit(s string) []string {
	return CamelSplit(s)
}

func PascalJoin(a []string) string {
	return UpperCamelJoin(a)
}
