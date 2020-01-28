package caseconv

import (
	"bytes"
	"strings"
)

var AcronymMap = map[string]bool{
	"HTTP": true,
	"ID":   true,
	"WWW":  true,
	"URL":  true,
	"DAO":  true,
	"XML":  true,
}

func IsAcronym(s string) bool {
	s = strings.ToUpper(s)
	_, ok := AcronymMap[s]
	return ok
}

func SimpleJoin(a []string, sep string, handle func(string) string) string {
	if len(a) == 0 {
		return ""
	}

	var buffer bytes.Buffer
	buffer.WriteString(handle(a[0]))
	for _, s := range a[1:] {
		buffer.WriteString(sep)
		buffer.WriteString(handle(s))
	}
	return buffer.String()
}
